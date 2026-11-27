/**
 * Photo gallery component — displays item photos with lightbox.
 */

import { element, on, removeElement, clearElement } from '../dom/element.js';
import { trapFocus, saveFocus, restoreFocus } from '../dom/focus.js';
import { deletePhoto, reorderPhotos, type Photo } from '../api/items.js';
import { createToast } from './toast.js';

interface PhotoGalleryOptions {
  photos: Photo[];
  itemId: string;
  baseVersion: number;
  onDelete: (photoId: string) => void;
  onReorder: (version: number, photos: Photo[]) => void;
}

export function createPhotoGallery(opts: PhotoGalleryOptions): HTMLElement {
  const { photos, itemId, baseVersion, onDelete, onReorder } = opts;
  const cleanupFns: (() => void)[] = [];
  let currentVersion = baseVersion;
  let currentPhotos: Photo[] = [...photos];

  const container = element({
    tag: 'div',
    className: 'photo-gallery',
    children: [
      element({ tag: 'h2', className: 'photo-gallery__title', children: ['Photos'] }),
      element({ tag: 'div', className: 'photo-gallery__grid' }),
      element({ tag: 'div', className: 'photo-gallery__toolbar' }),
    ],
  });

  const gridEl = container.querySelector('.photo-gallery__grid') as HTMLElement;
  const toolbarEl = container.querySelector('.photo-gallery__toolbar') as HTMLElement;

  // Render the gallery
  function render(): void {
    renderGrid();
    renderToolbar();
  }

  // Render photo grid
  function renderGrid(): void {
    clearElement(gridEl);

    if (currentPhotos.length === 0) {
      gridEl.appendChild(
        element({
          tag: 'p',
          className: 'photo-gallery__empty',
          children: ['No photos yet'],
        }),
      );
      return;
    }

    // Sort by sort_order
    const sorted = [...currentPhotos].sort((a, b) => a.sort_order - b.sort_order);

    for (const photo of sorted) {
      gridEl.appendChild(createPhotoCard(photo));
    }
  }

  // Create a single photo card
  function createPhotoCard(photo: Photo): HTMLElement {
    const card = element({
      tag: 'div',
      className: 'photo-gallery__card',
      attrs: {
        'data-photo-id': photo.id,
        'data-photo-version': String(photo.version),
      },
      children: [
        element({
          tag: 'img',
          attrs: {
            src: getPhotoThumbnailUrl(photo),
            alt: `Photo ${photo.sort_order + 1}`,
            loading: 'lazy',
            className: 'photo-gallery__thumb',
          },
        }),
        element({
          tag: 'div',
          className: 'photo-gallery__card-actions',
          children: [
            element({
              tag: 'button',
              attrs: {
                type: 'button',
                className: 'btn btn--icon',
                'aria-label': 'View full size',
                'data-action': 'view',
              },
              children: ['\u{1F50D}'], // magnifying glass
            }),
            element({
              tag: 'button',
              attrs: {
                type: 'button',
                className: 'btn btn--icon',
                'aria-label': 'Delete photo',
                'data-action': 'delete',
              },
              children: ['\u{2716}'], // X
            }),
          ],
        }),
      ],
    });

    // View button
    const viewBtn = card.querySelector('[data-action="view"]');
    if (viewBtn) {
      cleanupFns.push(on(viewBtn, 'click', () => openLightbox(photo)));
    }

    // Delete button
    const deleteBtn = card.querySelector('[data-action="delete"]');
    if (deleteBtn) {
      cleanupFns.push(on(deleteBtn, 'click', () => handleDelete(photo)));
    }

    return card;
  }

  // Render toolbar with reorder controls
  function renderToolbar(): void {
    clearElement(toolbarEl);

    if (currentPhotos.length <= 1) return;

    const toolbar = element({
      tag: 'div',
      className: 'photo-gallery__reorder',
      children: [
        element({
          tag: 'button',
          attrs: {
            type: 'button',
            className: 'btn btn--sm',
            'data-action': 'reorder-left',
            disabled: 'disabled',
          },
          children: ['\u2190 Left'],
        }),
        element({
          tag: 'button',
          attrs: {
            type: 'button',
            className: 'btn btn--sm',
            'data-action': 'reorder-right',
            disabled: 'disabled',
          },
          children: ['Right \u2192'],
        }),
      ],
    });

    toolbarEl.appendChild(toolbar);

    // Track selected photo for reorder
    let selectedPhotoId: string | null = null;

    // Click on card selects it for reorder
    cleanupFns.push(
      on(gridEl, 'click', (e: Event) => {
        const card = (e.target as HTMLElement).closest('.photo-gallery__card');
        if (!card) return;

        // Deselect previous
        const prevSelected = gridEl.querySelector('.photo-gallery__card--selected');
        if (prevSelected) {
          prevSelected.classList.remove('photo-gallery__card--selected');
        }

        const photoId = card.getAttribute('data-photo-id');
        if (photoId) {
          card.classList.add('photo-gallery__card--selected');
          selectedPhotoId = photoId;
          updateReorderButtons(selectedPhotoId);
        }
      }),
    );

    // Reorder left
    const leftBtn = toolbar.querySelector('[data-action="reorder-left"]');
    if (leftBtn) {
      cleanupFns.push(
        on(leftBtn, 'click', async () => {
          if (!selectedPhotoId) return;
          await handleReorder(selectedPhotoId, -1);
          updateReorderButtons(selectedPhotoId);
        }),
      );
    }

    // Reorder right
    const rightBtn = toolbar.querySelector('[data-action="reorder-right"]');
    if (rightBtn) {
      cleanupFns.push(
        on(rightBtn, 'click', async () => {
          if (!selectedPhotoId) return;
          await handleReorder(selectedPhotoId, 1);
          updateReorderButtons(selectedPhotoId);
        }),
      );
    }

    function updateReorderButtons(photoId: string): void {
      const sorted = [...currentPhotos].sort((a, b) => a.sort_order - b.sort_order);
      const index = sorted.findIndex((p) => p.id === photoId);

      if (leftBtn) {
        (leftBtn as HTMLButtonElement).disabled = index <= 0;
      }
      if (rightBtn) {
        (rightBtn as HTMLButtonElement).disabled = index >= sorted.length - 1;
      }
    }
  }

  // Handle photo deletion
  async function handleDelete(photo: Photo): Promise<void> {
    try {
      await deletePhoto(photo.id, photo.version);
      currentPhotos = currentPhotos.filter((p) => p.id !== photo.id);
      currentVersion++;
      onDelete(photo.id);
      createToast({ type: 'success', message: 'Photo deleted' });
      render();
    } catch (error: unknown) {
      if (error instanceof Error && 'kind' in error) {
        const apiError = error as Error & { kind: string; message: string };
        if (apiError.kind === 'version_conflict') {
          createToast({ type: 'error', message: 'Version conflict — reload and try again' });
        } else {
          createToast({ type: 'error', message: apiError.message });
        }
      } else {
        createToast({ type: 'error', message: 'Failed to delete photo' });
      }
    }
  }

  // Handle photo reordering
  async function handleReorder(photoId: string, direction: -1 | 1): Promise<void> {
    const sorted = [...currentPhotos].sort((a, b) => a.sort_order - b.sort_order);
    const index = sorted.findIndex((p) => p.id === photoId);
    if (index < 0 || index + direction < 0 || index + direction >= sorted.length) return;

    // Swap
    const swapIndex = index + direction;
    const photoA = sorted[index];
    const photoB = sorted[swapIndex];
    if (!photoA || !photoB) return;
    const tempOrder = photoA.sort_order;
    photoA.sort_order = photoB.sort_order;
    photoB.sort_order = tempOrder;

    // Send to API
    const orderedPhotos = sorted.map((p) => ({ id: p.id, version: p.version }));

    try {
      await reorderPhotos(itemId, currentVersion, orderedPhotos);
      currentPhotos = sorted;
      currentVersion++;
      onReorder(currentVersion, sorted);
      createToast({ type: 'success', message: 'Photo order updated' });
      render();
    } catch (error: unknown) {
      if (error instanceof Error && 'kind' in error) {
        const apiError = error as Error & { kind: string; message: string };
        if (apiError.kind === 'version_conflict') {
          createToast({ type: 'error', message: 'Version conflict — order may have changed' });
          // Reload to get latest order
          render();
        } else {
          createToast({ type: 'error', message: apiError.message });
        }
      } else {
        createToast({ type: 'error', message: 'Failed to reorder photos' });
      }
    }
  }

  // Get thumbnail URL for a photo
  function getPhotoThumbnailUrl(photo: Photo): string {
    // In production, this would be a thumbnail URL from the API
    // For now, use the full photo URL with a hint for the server
    return `/api/photos/${photo.id}/thumbnail`;
  }

  // Open lightbox with full-size photo
  function openLightbox(photo: Photo): void {
    saveFocus();

    const overlay = element({
      tag: 'div',
      attrs: {
        className: 'lightbox-overlay',
        'aria-modal': 'true',
        role: 'dialog',
        'aria-label': 'Photo viewer',
      },
    });

    const dialog = element({
      tag: 'div',
      className: 'lightbox',
      children: [
        element({
          tag: 'button',
          attrs: {
            type: 'button',
            className: 'lightbox__close',
            'aria-label': 'Close',
          },
          children: ['\u00d7'],
        }),
        element({
          tag: 'img',
          attrs: {
            src: `/api/photos/${photo.id}/original`,
            alt: `Photo ${photo.sort_order + 1}`,
            className: 'lightbox__image',
          },
        }),
        element({
          tag: 'div',
          className: 'lightbox__nav',
          children: [
            element({
              tag: 'button',
              attrs: {
                type: 'button',
                className: 'lightbox__prev',
                'aria-label': 'Previous photo',
              },
              children: ['\u2039'],
            }),
            element({
              tag: 'button',
              attrs: {
                type: 'button',
                className: 'lightbox__next',
                'aria-label': 'Next photo',
              },
              children: ['\u203A'],
            }),
          ],
        }),
      ],
    }) as HTMLDialogElement;

    overlay.appendChild(dialog);
    document.body.appendChild(overlay);

    dialog.showModal();
    const untrap = trapFocus(dialog);

    const sorted = [...currentPhotos].sort((a, b) => a.sort_order - b.sort_order);
    let currentIndex = sorted.findIndex((p) => p.id === photo.id);

    function closeLightbox(): void {
      untrap();
      document.body.removeChild(overlay);
      restoreFocus();
    }

    function navigate(direction: -1 | 1): void {
      const nextIndex = currentIndex + direction;
      if (nextIndex < 0 || nextIndex >= sorted.length) return;
      currentIndex = nextIndex;
      const img = dialog.querySelector('.lightbox__image');
      if (img) {
        const photo = sorted[currentIndex];
        if (photo) {
          (img as HTMLImageElement).src = `/api/photos/${photo.id}/original`;
        }
      }
    }

    const closeBtn = dialog.querySelector('.lightbox__close');
    if (closeBtn) {
      cleanupFns.push(on(closeBtn, 'click', closeLightbox));
    }
    cleanupFns.push(
      on(overlay, 'click', (e: Event) => {
        if ((e.target as HTMLElement) === overlay) closeLightbox();
      }),
    );

    const prevBtn = dialog.querySelector('.lightbox__prev');
    if (prevBtn) {
      cleanupFns.push(on(prevBtn, 'click', () => navigate(-1)));
    }

    const nextBtn = dialog.querySelector('.lightbox__next');
    if (nextBtn) {
      cleanupFns.push(on(nextBtn, 'click', () => navigate(1)));
    }

    // Keyboard navigation
    cleanupFns.push(
      on(document, 'keydown', ((e: Event) => {
        const ke = e as KeyboardEvent;
        if (ke.key === 'Escape') closeLightbox();
        if (ke.key === 'ArrowLeft') navigate(-1);
        if (ke.key === 'ArrowRight') navigate(1);
      }) as (e: Event) => void),
    );
  }

  // Public API to update photos
  function setPhotos(newPhotos: Photo[]): void {
    currentPhotos = [...newPhotos];
    currentVersion =
      newPhotos.length > 0 ? Math.max(...newPhotos.map((p) => p.version)) : currentVersion;
    render();
  }

  render();

  return {
    destroy: () => {
      for (const fn of cleanupFns) fn();
      removeElement(container);
    },
    getContainer: () => container,
    setPhotos,
  } as unknown as HTMLElement;
}
