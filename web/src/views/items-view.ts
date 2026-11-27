/**
 * Items list view — full implementation with search, filters, pagination, cards/table.
 */

import { element, on } from '../dom/element.js';
import type { View } from '../app/lifecycle.js';
import { navigate } from '../app/router.js';
import { fetchItems, type Item, type ItemsQueryParams } from '../api/items.js';
import { createItemCard } from '../components/item-card.js';
import { createItemFilters } from '../components/item-filters.js';
import { createLoadingState } from '../components/loading-state.js';
import { createEmptyState } from '../components/empty-state.js';
import { createErrorState } from '../components/error-state.js';
import { createConfirmDialog } from '../components/confirm-dialog.js';
import { serializeQueryParams, deserializeQueryParams } from '../domain/item.js';
import { normalizeApiError } from '../api/errors.js';

type ViewMode = 'cards' | 'table';

export class ItemsView implements View {
  private container: HTMLElement | null = null;
  private cleanup: (() => void)[] = [];
  private abortController: AbortController | null = null;
  private items: Item[] = [];
  private nextCursor: string | null = null;
  private queryParams: ItemsQueryParams = { sort: 'updated_desc', limit: 50 };
  private categories: { id: string; name: string }[] = [];
  private viewMode: ViewMode = this.loadViewMode();
  private isLoading = false;
  private lastError: Error | null = null;

  async mount(container: HTMLElement): Promise<void> {
    this.container = container;
    this.viewMode = this.loadViewMode();
    this.queryParams = this.loadQueryParams();
    document.title = 'Items — E-Wardrobe';

    await this.loadCategories();
    await this.loadItems();
  }

  destroy(): void {
    for (const fn of this.cleanup) {
      fn();
    }
    this.cleanup = [];
    if (this.abortController) {
      this.abortController.abort();
    }
    this.container = null;
  }

  private loadViewMode(): ViewMode {
    try {
      const stored = localStorage.getItem('wardrobe:list_mode');
      return stored === 'table' || stored === 'cards' ? stored : 'cards';
    } catch {
      return 'cards';
    }
  }

  private saveViewMode(mode: ViewMode): void {
    try {
      localStorage.setItem('wardrobe:list_mode', mode);
    } catch {
      // Storage unavailable
    }
  }

  private loadQueryParams(): ItemsQueryParams {
    const params = deserializeQueryParams(window.location.search);
    const result: ItemsQueryParams = { sort: 'updated_desc', limit: 50 };

    if (params.q) result.q = String(params.q);
    if (params.category_id) result.category_id = String(params.category_id);
    if (params.status) result.status = String(params.status) as ItemsQueryParams['status'];
    if (params.season) result.season = String(params.season) as ItemsQueryParams['season'];
    if (params.condition_rating) result.condition_rating = Number(params.condition_rating);
    if (params.preference_rating) result.preference_rating = Number(params.preference_rating);
    if (params.archived) result.archived = Boolean(params.archived);
    if (params.sort) result.sort = String(params.sort);
    if (params.cursor) result.cursor = String(params.cursor);
    if (params.limit) result.limit = Number(params.limit);

    return result;
  }

  private updateUrl(): void {
    const params = serializeQueryParams(this.queryParams);
    const search = new URLSearchParams(params).toString();
    const newUrl = `${window.location.pathname}${search ? '?' + search : ''}`;
    window.history.replaceState({}, '', newUrl);
  }

  private async loadCategories(): Promise<void> {
    try {
      const resp = await fetch('/api/v1/categories', { credentials: 'same-origin' });
      if (!resp.ok) return;
      const data = await resp.json();
      if (data && Array.isArray(data.categories)) {
        this.categories = data.categories.map((c: { id: string; name: string }) => ({
          id: c.id,
          name: c.name,
        }));
      }
    } catch {
      // Categories not critical for display
    }
  }

  private async loadItems(): Promise<void> {
    if (this.isLoading) return;
    this.isLoading = true;

    if (this.abortController) {
      this.abortController.abort();
    }
    this.abortController = new AbortController();

    this.render();

    try {
      const data = await fetchItems(this.queryParams);
      this.items = data.items;
      this.nextCursor = data.next_cursor;
      this.lastError = null;
    } catch (error: unknown) {
      if (error instanceof DOMException && error.name === 'AbortError') return;
      this.lastError = error instanceof Error ? error : new Error(String(error));
    } finally {
      this.isLoading = false;
    }

    this.render();
  }

  private render(): void {
    const container = this.container;
    if (!container) return;

    container.innerHTML = '';

    container.appendChild(
      element({
        tag: 'div',
        className: 'view-header',
        children: [
          element({
            tag: 'h1',
            attrs: { class: 'view-title' },
            children: ['Items'],
          }),
          element({
            tag: 'button',
            attrs: { type: 'button', className: 'btn btn--primary', 'data-action': 'add' },
            children: ['Add Item'],
          }),
        ],
      }),
    );

    const filters = createItemFilters({
      value: this.queryParams,
      onChange: (params) => {
        this.queryParams = params;
        this.updateUrl();
        void this.loadItems();
      },
      categories: this.categories,
    });
    container.appendChild(filters);

    const toggleContainer = element({
      tag: 'div',
      className: 'view-toggle',
      children: [
        element({
          tag: 'button',
          attrs: {
            type: 'button',
            className: `btn btn--sm ${this.viewMode === 'cards' ? 'btn--primary' : ''}`,
            'data-action': 'mode-cards',
          },
          children: ['Cards'],
        }),
        element({
          tag: 'button',
          attrs: {
            type: 'button',
            className: `btn btn--sm ${this.viewMode === 'table' ? 'btn--primary' : ''}`,
            'data-action': 'mode-table',
          },
          children: ['Table'],
        }),
      ],
    });
    container.appendChild(toggleContainer);

    toggleContainer.addEventListener('click', (e: Event) => {
      const target = e.target as HTMLElement;
      const action = target.getAttribute('data-action');
      if (action === 'mode-cards') {
        this.viewMode = 'cards';
        this.saveViewMode('cards');
        this.render();
      } else if (action === 'mode-table') {
        this.viewMode = 'table';
        this.saveViewMode('table');
        this.render();
      }
    });

    const mainContent = element({ tag: 'main-content', className: 'items-list' });

    if (this.isLoading) {
      mainContent.appendChild(createLoadingState('Loading items...'));
    } else if (this.lastError) {
      const apiError = normalizeApiError(null, null, true);
      mainContent.appendChild(
        createErrorState({
          title: apiError.httpStatus === 0 ? 'Connection error' : 'Error',
          message: this.lastError.message || 'Failed to load items',
          retryLabel: 'Retry',
          onRetry: () => void this.loadItems(),
        }),
      );
    } else if (this.items.length === 0) {
      mainContent.appendChild(
        createEmptyState({
          title: 'No items found',
          description: 'Add your first item to get started.',
          actionLabel: 'Add Item',
          onAction: () => navigate('/items/new'),
        }),
      );
    } else {
      if (this.viewMode === 'cards') {
        for (const item of this.items) {
          mainContent.appendChild(
            createItemCard({
              item,
              onEdit: (id) => navigate(`/items/${id}/edit`),
              onDelete: async () => {
                const confirmed = await createConfirmDialog({
                  title: 'Delete item',
                  message: 'Are you sure you want to delete this item? This action can be undone.',
                  confirmLabel: 'Delete',
                  cancelLabel: 'Cancel',
                  confirmClass: 'btn--danger',
                });

                if (confirmed) {
                  await this.loadItems();
                }
              },
            }),
          );
        }
      } else {
        mainContent.appendChild(this.renderTable());
      }
    }

    container.appendChild(mainContent);

    if (this.nextCursor) {
      const pagination = element({
        tag: 'div',
        className: 'pagination',
        children: [
          element({
            tag: 'button',
            attrs: { type: 'button', className: 'btn btn--sm', 'data-action': 'load-more' },
            children: ['Load more'],
          }),
        ],
      });

      pagination.addEventListener('click', (e: Event) => {
        if ((e.target as HTMLElement).getAttribute('data-action') === 'load-more') {
          this.queryParams.cursor = this.nextCursor || undefined;
          void this.loadItems();
        }
      });

      container.appendChild(pagination);
    }

    const addBtn = container.querySelector('[data-action="add"]');
    if (addBtn) {
      this.cleanup.push(on(addBtn, 'click', () => navigate('/items/new')));
    }
  }

  private renderTable(): HTMLElement {
    const table = element({
      tag: 'table',
      className: 'items-table',
      attrs: { role: 'grid' },
      children: [
        element({
          tag: 'thead',
          children: [
            element({
              tag: 'tr',
              children: [
                element({ tag: 'th', attrs: { scope: 'col' }, children: ['Name'] }),
                element({ tag: 'th', attrs: { scope: 'col' }, children: ['Category'] }),
                element({ tag: 'th', attrs: { scope: 'col' }, children: ['Brand'] }),
                element({ tag: 'th', attrs: { scope: 'col' }, children: ['Price'] }),
                element({ tag: 'th', attrs: { scope: 'col' }, children: ['Status'] }),
                element({ tag: 'th', attrs: { scope: 'col' }, children: ['Actions'] }),
              ],
            }),
          ],
        }),
        element({
          tag: 'tbody',
          children: this.items.map((item) =>
            element({
              tag: 'tr',
              attrs: { 'data-item-id': item.id, role: 'row' },
              children: [
                element({ tag: 'td', children: [item.name] }),
                element({ tag: 'td', children: ['—'] }),
                element({ tag: 'td', children: [item.brand || '—'] }),
                element({
                  tag: 'td',
                  children: [
                    item.purchase_price
                      ? `${item.purchase_price} ${item.price_currency ?? ''}`
                      : '—',
                  ],
                }),
                element({ tag: 'td', children: [item.status] }),
                element({
                  tag: 'td',
                  children: [
                    element({
                      tag: 'button',
                      attrs: {
                        type: 'button',
                        className: 'btn btn--icon',
                        'data-action': 'edit',
                        'data-item-id': item.id,
                      },
                      children: ['Edit'],
                    }),
                  ],
                }),
              ],
            }),
          ),
        }),
      ],
    });

    return table;
  }
}
