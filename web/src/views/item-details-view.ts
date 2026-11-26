/**
 * Item details view — shows full item info with actions.
 */

import { element, on } from '../dom/element.js';
import type { View } from '../app/lifecycle.js';
import { navigate, getCurrentParams } from '../app/router.js';
import { fetchItem, updateItem, archiveItem, deleteItem, type Item } from '../api/items.js';
import { createConfirmDialog } from '../components/confirm-dialog.js';
import { createConflictDialog } from '../components/conflict-dialog.js';
import { formatPrice } from '../domain/item.js';
import { createToast } from '../components/toast.js';
import { createPhotoUploader } from '../components/photo-uploader.js';
import { createPhotoGallery } from '../components/photo-gallery.js';

export class ItemDetailsView implements View {
  private container: HTMLElement | null = null;
  private cleanup: (() => void)[] = [];
  private abortController: AbortController | null = null;
  private item: Item | null = null;
  private photoUploader: HTMLElement | null = null;
  private photoGallery: HTMLElement | null = null;

  async mount(container: HTMLElement): Promise<void> {
    this.container = container;
    document.title = 'Item — E-Wardrobe';
    await this.loadItem();
  }

  destroy(): void {
    for (const fn of this.cleanup) {
      fn();
    }
    this.cleanup = [];
    if (this.abortController) {
      this.abortController.abort();
    }
    if (this.photoUploader) {
      (this.photoUploader as unknown as { destroy?: () => void }).destroy?.();
      this.photoUploader = null;
    }
    if (this.photoGallery) {
      (this.photoGallery as unknown as { destroy?: () => void }).destroy?.();
      this.photoGallery = null;
    }
    this.container = null;
  }

  private async loadItem(): Promise<void> {
    if (this.abortController) {
      this.abortController.abort();
    }
    this.abortController = new AbortController();

    const params = getCurrentParams();
    const itemId = params.itemId;
    if (!itemId) {
      this.renderNotFound();
      return;
    }

    this.renderLoading();

    try {
      this.item = await fetchItem(itemId);
      this.render();
    } catch (error: unknown) {
      if (error instanceof DOMException && error.name === 'AbortError') return;
      this.renderError(error);
    }
  }

  private renderLoading(): void {
    const container = this.container;
    if (!container) return;
    container.innerHTML = '';
    container.appendChild(
      element({
        tag: 'div',
        className: 'loading-state',
        children: ['Loading item...'],
      }),
    );
  }

  private renderNotFound(): void {
    const container = this.container;
    if (!container) return;
    container.innerHTML = '';
    container.appendChild(
      element({
        tag: 'h1',
        attrs: { class: 'view-title' },
        children: ['Item not found'],
      }),
    );
  }

  private renderError(error: unknown): void {
    const container = this.container;
    if (!container) return;
    container.innerHTML = '';
    const message = error instanceof Error ? error.message : 'Failed to load item';
    container.appendChild(
      element({
        tag: 'div',
        className: 'error-state',
        children: [
          element({ tag: 'h2', children: ['Error'] }),
          element({ tag: 'p', children: [message] }),
        ],
      }),
    );
  }

  private render(): void {
    const container = this.container;
    if (!container || !this.item) return;

    const current = this.item;
    container.innerHTML = '';

    // Header with actions
    container.appendChild(
      element({
        tag: 'div',
        className: 'view-header',
        children: [
          element({
            tag: 'h1',
            attrs: { class: 'view-title' },
            children: [current.name],
          }),
          element({
            tag: 'div',
            className: 'view-actions',
            children: [
              element({
                tag: 'button',
                attrs: { type: 'button', className: 'btn', 'data-action': 'edit' },
                children: ['Edit'],
              }),
              element({
                tag: 'button',
                attrs: {
                  type: 'button',
                  className: `btn ${current.archived_at ? 'btn--primary' : 'btn--danger'}`,
                  'data-action': 'archive',
                },
                children: [this.item.archived_at ? 'Unarchive' : 'Archive'],
              }),
              element({
                tag: 'button',
                attrs: { type: 'button', className: 'btn btn--danger', 'data-action': 'delete' },
                children: ['Delete'],
              }),
            ],
          }),
        ],
      }),
    );

    // Detail fields
    const details = element({
      tag: 'dl',
      className: 'item-details',
      children: [
        this.field('Brand', this.item.brand),
        this.field('Category', '—'),
        this.field('Season', this.item.season?.replace('_', ' ') ?? null),
        this.field('Color', this.item.color),
        this.field('Size', this.item.size),
        this.field('Status', this.item.status),
        this.field(
          'Condition',
          this.item.condition_rating ? `${this.item.condition_rating}/5` : null,
        ),
        this.field(
          'Preference',
          this.item.preference_rating ? `${this.item.preference_rating}/5` : null,
        ),
        this.field('Purchase Year', this.item.purchase_year?.toString() ?? null),
        this.field('Price', formatPrice(this.item.purchase_price, this.item.price_currency)),
        this.field('Purchase Place', this.item.purchase_place),
        this.field('Last Worn', this.item.last_worn_on),
        this.field('Storage', this.item.storage_location),
        this.field('Description', this.item.description),
      ].filter(Boolean),
    });

    container.appendChild(details);

    // Quick actions
    const quickActions = element({
      tag: 'div',
      className: 'quick-actions',
      children: [
        element({
          tag: 'button',
          attrs: { type: 'button', className: 'btn btn--sm', 'data-action': 'worn-today' },
          children: ['Worn today'],
        }),
      ],
    });

    container.appendChild(quickActions);

    // Photo section with uploader and gallery
    const photoSection = element({
      tag: 'section',
      className: 'photo-section',
      children: [
        element({ tag: 'h2', className: 'photo-section__title', children: ['Photos'] }),
      ],
    });

    container.appendChild(photoSection);

    // Create photo gallery
    this.photoGallery = createPhotoGallery({
      photos: this.item.photos,
      itemId: this.item.id,
      baseVersion: this.item.version,
      onDelete: () => { /* gallery handles its own DOM */ },
      onReorder: (version, photos) => {
        this.item = this.item ? { ...this.item, version, photos } : null;
      },
    }) as HTMLElement;

    photoSection.appendChild(this.photoGallery);

    // Create photo uploader
    this.photoUploader = createPhotoUploader({
      itemId: this.item.id,
      onUploadComplete: () => {
        void this.loadItem();
      },
      onUploadError: (message) => {
        createToast({ type: 'error', message });
      },
    }) as HTMLElement;

    photoSection.appendChild(this.photoUploader);

    // Event handlers
    const editBtn = container.querySelector('[data-action="edit"]');
    if (editBtn) {
      this.cleanup.push(on(editBtn, 'click', () => navigate(`/items/${current.id}/edit`)));
    }

    const archiveBtn = container.querySelector('[data-action="archive"]');
    if (archiveBtn) {
      this.cleanup.push(
        on(archiveBtn, 'click', async () => {
          try {
            const newArchived = !current.archived_at;
            await archiveItem(current.id, {
              base_version: current.version,
              archived: newArchived,
            });
            createToast({
              type: 'success',
              message: newArchived ? 'Item archived' : 'Item unarchived',
            });
            await this.loadItem();
          } catch (error: unknown) {
            if (error instanceof Error && 'kind' in error) {
              const apiError = error as Error & { kind: string; message: string };
              if (apiError.kind === 'version_conflict') {
                createConflictDialog({
                  original: current,
                  draft: { storage_location: 'Updated' },
                  onUseServer: () => this.loadItem(),
                  onKeepDraft: () => {},
                  onManual: () => {},
                });
              } else {
                createToast({ type: 'error', message: apiError.message });
              }
            }
          }
        }),
      );
    }

    const deleteBtn = container.querySelector('[data-action="delete"]');
    if (deleteBtn) {
      this.cleanup.push(
        on(deleteBtn, 'click', async () => {
          const confirmed = await createConfirmDialog({
            title: 'Delete item',
            message: 'Are you sure you want to delete this item? This cannot be undone.',
            confirmLabel: 'Delete',
            confirmClass: 'btn--danger',
          });

          if (confirmed && this.item) {
            try {
              await deleteItem(this.item.id, this.item.version);
              createToast({ type: 'success', message: 'Item deleted' });
              navigate('/items');
            } catch (error: unknown) {
              if (error instanceof Error && 'kind' in error) {
                const apiError = error as Error & { kind: string; message: string };
                createToast({ type: 'error', message: apiError.message });
              }
            }
          }
        }),
      );
    }

    const wornBtn = container.querySelector('[data-action="worn-today"]');
    if (wornBtn) {
      this.cleanup.push(
        on(wornBtn, 'click', async () => {
          if (!this.item) return;
          try {
            const today = new Date().toISOString().split('T')[0];
            await updateItem(this.item.id, {
              base_version: this.item.version,
              fields: { last_worn_on: today },
            });
            createToast({ type: 'success', message: 'Marked as worn today' });
            await this.loadItem();
          } catch (error: unknown) {
            if (error instanceof Error && 'kind' in error) {
              const apiError = error as Error & { kind: string; message: string };
              createToast({ type: 'error', message: apiError.message });
            }
          }
        }),
      );
    }
  }

  private field(label: string, value: string | null): HTMLElement | null {
    if (!value) return null;
    return element({
      tag: 'div',
      className: 'item-details__field',
      children: [
        element({ tag: 'dt', className: 'item-details__label', children: [label] }),
        element({ tag: 'dd', className: 'item-details__value', children: [value] }),
      ],
    });
  }
}
