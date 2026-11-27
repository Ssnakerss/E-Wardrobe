/**
 * Item form view — create and edit items with full field support.
 */

import { element, on, clearElement } from '../dom/element.js';
import type { View } from '../app/lifecycle.js';
import { navigate, getCurrentParams } from '../app/router.js';
import {
  fetchItem,
  createItem,
  updateItem,
  fetchStorageLocations,
  type Item,
  type ItemInput,
  type ItemFields,
  type Season,
  type ItemStatus,
} from '../api/items.js';
import { buildItemPatch, parsePrice, validateItemForm } from '../domain/item.js';
import { createToast } from '../components/toast.js';

interface FormField {
  name: keyof ItemFields;
  label: string;
  type: 'text' | 'select' | 'number' | 'date' | 'textarea';
  required?: boolean;
  options?: { value: string; label: string }[];
}

const SEASONS: { value: Season; label: string }[] = [
  { value: 'winter', label: 'Winter' },
  { value: 'spring', label: 'Spring' },
  { value: 'summer', label: 'Summer' },
  { value: 'autumn', label: 'Autumn' },
  { value: 'demi_season', label: 'Demi-season' },
  { value: 'all_season', label: 'All season' },
];

const STATUSES: { value: ItemStatus; label: string }[] = [
  { value: 'in_use', label: 'In use' },
  { value: 'stored', label: 'Stored' },
  { value: 'needs_repair', label: 'Needs repair' },
];

const CURRENCIES = ['RUB', 'USD', 'EUR', 'KZT', 'CNY'];

const FIELDS: FormField[] = [
  { name: 'name', label: 'Name', type: 'text', required: true },
  { name: 'brand', label: 'Brand', type: 'text' },
  { name: 'description', label: 'Description', type: 'textarea' },
  { name: 'purchase_year', label: 'Purchase Year', type: 'number' },
  { name: 'purchase_price', label: 'Price', type: 'text' },
  {
    name: 'price_currency',
    label: 'Currency',
    type: 'select',
    options: CURRENCIES.map((c) => ({ value: c, label: c })),
  },
  { name: 'purchase_place', label: 'Purchase Place', type: 'text' },
  {
    name: 'condition_rating',
    label: 'Condition',
    type: 'select',
    options: [1, 2, 3, 4, 5].map((v) => ({ value: String(v), label: String(v) })),
  },
  {
    name: 'preference_rating',
    label: 'Preference',
    type: 'select',
    options: [1, 2, 3, 4, 5].map((v) => ({ value: String(v), label: String(v) })),
  },
  { name: 'last_worn_on', label: 'Last Worn', type: 'date' },
  { name: 'storage_location', label: 'Storage Location', type: 'text' },
  { name: 'season', label: 'Season', type: 'select', options: SEASONS },
  { name: 'color', label: 'Color', type: 'text' },
  { name: 'size', label: 'Size', type: 'text' },
  { name: 'status', label: 'Status', type: 'select', options: STATUSES },
  { name: 'category_id', label: 'Category', type: 'select', required: true },
];

export class ItemFormView implements View {
  private container: HTMLElement | null = null;
  private cleanup: (() => void)[] = [];
  private abortController: AbortController | null = null;
  private item: Item | null = null;
  private isEditMode = false;
  private isLoading = false;
  private categories: { id: string; name: string }[] = [];
  private storageSuggestions: string[] = [];
  private storageDebounceTimer: ReturnType<typeof setTimeout> | null = null;
  private _loadError: string | null = null;

  async mount(container: HTMLElement): Promise<void> {
    this.container = container;
    this.isEditMode = this.checkEditMode();
    document.title = this.isEditMode ? 'Edit Item — E-Wardrobe' : 'New Item — E-Wardrobe';

    await this.loadCategories();

    if (this.isEditMode) {
      await this.loadItem();
    }

    this.render();
  }

  destroy(): void {
    for (const fn of this.cleanup) {
      fn();
    }
    this.cleanup = [];
    if (this.abortController) {
      this.abortController.abort();
    }
    if (this.storageDebounceTimer) {
      clearTimeout(this.storageDebounceTimer);
    }
    this.container = null;
  }

  private checkEditMode(): boolean {
    const path = window.location.pathname;
    return path.includes('/edit');
  }

  private async loadCategories(): Promise<void> {
    try {
      const resp = await fetch('/api/categories', { credentials: 'same-origin' });
      if (!resp.ok) return;
      const data = await resp.json();
      if (data && Array.isArray(data.categories)) {
        this.categories = data.categories.map((c: { id: string; name: string }) => ({
          id: c.id,
          name: c.name,
        }));
      }
    } catch {
      // Non-critical
    }
  }

  private async loadItem(): Promise<void> {
    if (this.abortController) {
      this.abortController.abort();
    }
    this.abortController = new AbortController();

    const params = getCurrentParams();
    const itemId = params.itemId;
    if (!itemId) return;

    try {
      this.item = await fetchItem(itemId);
    } catch {
      this.item = null;
      this._loadError = 'Item not found';
    }
  }

  private render(): void {
    const container = this.container;
    if (!container) return;

    container.innerHTML = '';

    // Handle load error (item not found in edit mode)
    if (this._loadError) {
      container.appendChild(
        element({
          tag: 'div',
          className: 'error-state',
          children: [
            element({
              tag: 'h2',
              className: 'error-state__title',
              children: ['Item not found'],
            }),
            element({
              tag: 'p',
              className: 'error-state__message',
              children: [this._loadError],
            }),
            element({
              tag: 'a',
              attrs: { href: '/items', className: 'btn btn--primary' },
              children: ['Back to items'],
            }),
          ],
        }),
      );
      return;
    }

    container.appendChild(
      element({
        tag: 'h1',
        attrs: { class: 'view-title' },
        children: [this.isEditMode ? 'Edit Item' : 'New Item'],
      }),
    );

    const form = element({
      tag: 'form',
      attrs: { id: 'item-form', noValidate: true },
      children: this.buildFormFields(),
    });

    container.appendChild(form);

    // Error container
    const errorContainer = element({
      tag: 'div',
      attrs: { id: 'form-error', className: 'form-error', role: 'alert' },
    });
    container.appendChild(errorContainer);

    // Actions
    container.appendChild(
      element({
        tag: 'div',
        attrs: { className: 'form-actions' },
        children: [
          element({
            tag: 'button',
            attrs: { type: 'submit', className: 'btn btn--primary', id: 'form-submit' },
            children: [this.isEditMode ? 'Save Changes' : 'Create Item'],
          }),
          element({
            tag: 'a',
            attrs: { href: '/items', className: 'btn' },
            children: ['Cancel'],
          }),
        ],
      }),
    );

    // Submit handler
    const submitHandler = async (e: Event) => {
      e.preventDefault();
      if (this.isLoading) return;

      const formEl = container.querySelector<HTMLFormElement>('#item-form');
      if (!formEl) return;

      if (!formEl.checkValidity()) {
        formEl.reportValidity();
        return;
      }

      const errorContainerEl = container.querySelector<HTMLDivElement>('#form-error');
      if (!errorContainerEl) return;

      // Collect form data
      const draft = this.collectFormData(formEl);

      // Domain validation
      const validationErrors = validateItemForm(draft as Partial<ItemInput>);
      if (Object.keys(validationErrors).length > 0) {
        errorContainerEl.textContent = Object.values(validationErrors).join('; ');
        return;
      }

      this.isLoading = true;
      errorContainerEl.textContent = '';

      const submitBtn = formEl.querySelector<HTMLButtonElement>('#form-submit');
      const originalText = submitBtn?.textContent;
      if (submitBtn) {
        submitBtn.disabled = true;
        submitBtn.textContent = 'Saving...';
      }

      try {
        if (this.isEditMode && this.item) {
          const patch = buildItemPatch(this.item, draft);
          if (Object.keys(patch.fields).length === 0) {
            createToast({ type: 'info', message: 'No changes to save' });
            return;
          }
          const updated = await updateItem(this.item.id, patch);
          this.item = updated;
          createToast({ type: 'success', message: 'Item updated' });
          navigate(`/items/${this.item.id}`);
        } else {
          const input = draft as ItemInput;
          if (!input.status) input.status = 'in_use';
          if (!input.category_id) {
            errorContainerEl.textContent = 'Category is required';
            return;
          }
          const created = await createItem(input);
          createToast({ type: 'success', message: 'Item created' });
          navigate(`/items/${created.id}`);
        }
      } catch (error: unknown) {
        if (error instanceof Error && 'kind' in error) {
          const apiError = error as Error & {
            kind: string;
            message: string;
            details?: Record<string, string[]>;
          };
          if (apiError.kind === 'validation' && apiError.details) {
            const msgs = Object.values(apiError.details).flat();
            errorContainerEl.textContent = msgs.join('; ');
          } else {
            errorContainerEl.textContent = apiError.message || 'Save failed';
          }
        } else {
          errorContainerEl.textContent = 'Network error';
        }
      } finally {
        this.isLoading = false;
        if (submitBtn) {
          submitBtn.disabled = false;
          submitBtn.textContent =
            originalText ?? (this.isEditMode ? 'Save Changes' : 'Create Item');
        }
      }
    };

    this.cleanup.push(on(form, 'submit', submitHandler));

    // Storage location autocomplete
    const storageInput = container.querySelector<HTMLInputElement>('#storage_location');
    if (storageInput) {
      this.cleanup.push(
        on(storageInput, 'input', () => {
          if (this.storageDebounceTimer) {
            clearTimeout(this.storageDebounceTimer);
          }
          this.storageDebounceTimer = setTimeout(async () => {
            const value = storageInput.value;
            if (value.length < 2) return;
            try {
              const data = await fetchStorageLocations(value, 10);
              this.storageSuggestions = data.suggestions || [];
              this.populateDatalist(this.storageSuggestions);
            } catch {
              // Non-critical
            }
          }, 300);
        }),
      );
    }
  }

  private buildFormFields(): (Node | string)[] {
    const fields: (Node | string)[] = [];

    for (const field of FIELDS) {
      const group = element({
        tag: 'div',
        attrs: { className: 'form-group' },
        children: [
          element({
            tag: 'label',
            attrs: { htmlFor: field.name },
            children: [`${field.label}${field.required ? ' *' : ''}`],
          }),
        ],
      });

      let input: HTMLElement;

      if (field.type === 'select' && field.options) {
        const options = [
          element({ tag: 'option', attrs: { value: '' }, children: ['— Select —'] }),
        ];
        for (const opt of field.options) {
          options.push(
            element({
              tag: 'option',
              attrs: { value: opt.value },
              children: [opt.label],
            }),
          );
        }
        input = element({
          tag: 'select',
          attrs: { id: field.name, name: field.name },
          children: options,
        });
      } else if (field.type === 'textarea') {
        input = element({
          tag: 'textarea',
          attrs: {
            id: field.name,
            name: field.name,
            rows: '3',
            ...(field.required && { required: true }),
          },
        });
      } else {
        input = element({
          tag: 'input',
          attrs: {
            id: field.name,
            name: field.name,
            type: field.type === 'number' ? 'number' : field.type === 'date' ? 'date' : 'text',
            ...(field.required && { required: true }),
          },
        });
      }

      // Set default value for edit mode
      if (this.isEditMode && this.item && (field.name as string) in this.item) {
        const val = (this.item as unknown as Record<string, unknown>)[field.name];
        if (val !== null && val !== undefined && String(val) !== '') {
          (input as HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement).value = String(val);
        }
      }

      // Category select needs dynamic options
      if (field.name === 'category_id' && this.categories.length > 0) {
        (input as HTMLSelectElement).innerHTML = '';
        (input as HTMLSelectElement).appendChild(
          element({
            tag: 'option',
            attrs: { value: '' },
            children: ['— Select category —'],
          }),
        );
        for (const cat of this.categories) {
          const isSelected = this.isEditMode && this.item?.category_id === cat.id;
          const opt = element({
            tag: 'option',
            attrs: { value: cat.id, selected: isSelected ? '' : null },
            children: [cat.name],
          });
          (input as HTMLSelectElement).appendChild(opt);
        }
      }

      group.appendChild(input);
      fields.push(group);
    }

    return fields;
  }

  private collectFormData(formEl: HTMLFormElement): Partial<ItemFields> {
    const data: Partial<ItemFields> = {};

    for (const field of FIELDS) {
      const input = formEl.querySelector<
        HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement
      >(`[name="${field.name}"]`);
      if (!input) continue;

      let value: string | number | null = input.value;

      // Handle empty values
      if (value === '' || value === null) {
        // For numeric/date fields, set null
        if (field.type === 'number' || field.type === 'date') {
          (data as Record<string, unknown>)[field.name] = null;
        }
        continue;
      }

      // Parse price
      if (field.name === 'purchase_price') {
        const parsed = parsePrice(value);
        data[field.name] = parsed;
        continue;
      }

      // Parse numeric fields
      if (field.type === 'number') {
        const num = parseInt(value, 10);
        (data as Record<string, unknown>)[field.name] = isNaN(num) ? null : num;
        continue;
      }

      // Set value
      (data as Record<string, unknown>)[field.name] = value;
    }

    return data;
  }

  private populateDatalist(suggestions: string[]): void {
    // Simple approach: use a datalist element
    let datalist = document.getElementById('storage-suggestions');
    if (!datalist) {
      datalist = element({ tag: 'datalist', attrs: { id: 'storage-suggestions' } });
      document.body.appendChild(datalist);
    }

    clearElement(datalist);

    for (const suggestion of suggestions) {
      datalist.appendChild(
        element({
          tag: 'option',
          attrs: { value: suggestion },
          children: [suggestion],
        }),
      );
    }

    const storageInput = document.querySelector<HTMLInputElement>('#storage_location');
    if (storageInput) {
      storageInput.setAttribute('list', 'storage-suggestions');
    }
  }
}
