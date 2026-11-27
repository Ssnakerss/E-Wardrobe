/**
 * Photo uploader component — drag & drop + file picker with queue management.
 */

import { element, on, removeElement } from '../dom/element.js';
import { PhotoUploadQueue, type UploadFile } from '../domain/photo.js';
import { uploadPhoto, type Photo } from '../api/items.js';

export type PhotoUploaderStatus = 'idle' | 'uploading' | 'ready' | 'failed';

interface PhotoUploaderOptions {
  itemId: string;
  onUploadComplete: (photos: Photo[]) => void;
  onUploadError: (message: string) => void;
}

export function createPhotoUploader(opts: PhotoUploaderOptions): HTMLElement {
  const { itemId, onUploadComplete, onUploadError } = opts;
  const queue = new PhotoUploadQueue({ maxConcurrency: 3 });
  const cleanupFns: (() => void)[] = [];
  let destroyed = false;

  const container = element({
    tag: 'div',
    className: 'photo-uploader',
    children: [
      element({
        tag: 'div',
        className: 'photo-uploader__dropzone',
        attrs: {
          role: 'button',
          tabindex: '0',
          'aria-label': 'Upload photos',
        },
        children: [
          element({
            tag: 'input',
            attrs: {
              type: 'file',
              className: 'photo-uploader__input',
              accept: 'image/jpeg',
              multiple: 'multiple',
            },
          }),
          element({
            tag: 'p',
            className: 'photo-uploader__hint',
            children: ['Drop JPEG files here or click to browse'],
          }),
          element({
            tag: 'p',
            className: 'photo-uploader__hint',
            children: ['Max 12 MiB per file'],
          }),
        ],
      }),
      element({ tag: 'div', className: 'photo-uploader__queue' }),
    ],
  });

  const dropzone = container.querySelector('.photo-uploader__dropzone') as HTMLElement;
  const fileInput = container.querySelector('.photo-uploader__input') as HTMLInputElement;
  const queueEl = container.querySelector('.photo-uploader__queue') as HTMLElement;

  // Handle file selection
  function handleFiles(files: FileList | File[]): void {
    const jpegFiles: File[] = [];
    for (const file of files) {
      const uploaded = queue.enqueue(file, itemId);
      if (uploaded) {
        jpegFiles.push(file);
      }
    }

    if (jpegFiles.length > 0) {
      void startUpload();
    }
  }

  // Start upload of all queued files
  async function startUpload(): Promise<void> {
    await queue.upload(itemId, async (_fileId, file, sortOrder) => {
      try {
        const response = await uploadPhoto(itemId, file, sortOrder);
        return response as unknown as Response;
      } catch (error) {
        if (error instanceof Error && 'kind' in error) {
          throw error;
        }
        throw error;
      }
    });

    if (destroyed) return;

    // Check for failures
    const failed = queue.getByStatus('failed');
    if (failed.length > 0) {
      onUploadError(`${failed.length} photo(s) failed to upload`);
    }

    // Notify with current item photos
    onUploadComplete([]);
  }

  // Render the upload queue
  function renderQueue(): void {
    queueEl.innerHTML = '';
    const files = queue.files;

    for (const uploadFile of files) {
      queueEl.appendChild(createQueueItem(uploadFile, queue));
    }
  }

  // Create a single queue item element
  function createQueueItem(uploadFile: UploadFile, q: PhotoUploadQueue): HTMLElement {
    const item = element({
      tag: 'div',
      className: `photo-uploader__item photo-uploader__item--${uploadFile.status}`,
      children: [
        element({
          tag: 'span',
          className: 'photo-uploader__item-name',
          children: [uploadFile.file.name],
        }),
        element({
          tag: 'span',
          className: `photo-uploader__item-status photo-uploader__badge--${uploadFile.status}`,
          children: [formatStatus(uploadFile.status)],
        }),
      ],
    });

    // Add retry button for failed items
    if (uploadFile.status === 'failed') {
      const retryBtn = element({
        tag: 'button',
        attrs: {
          type: 'button',
          className: 'btn btn--sm',
          'aria-label': 'Retry upload',
        },
        children: ['Retry'],
      });

      item.appendChild(retryBtn);
      cleanupFns.push(
        on(retryBtn, 'click', async () => {
          try {
            await q.retry(uploadFile.id, itemId, async (_fileId, file, sortOrder) => {
              const response = await uploadPhoto(itemId, file, sortOrder);
              return response as unknown as Response;
            });
            renderQueue();
          } catch {
            onUploadError('Retry failed');
          }
        }),
      );
    }

    // Remove completed items after a delay
    if (uploadFile.status === 'ready') {
      setTimeout(() => {
        if (!destroyed) {
          const el = item.closest('.photo-uploader__item');
          if (el) removeElement(el as HTMLElement);
        }
      }, 3000);
    }

    return item;
  }

  function formatStatus(status: string): string {
    switch (status) {
      case 'queued':
        return 'Queued';
      case 'uploading':
        return 'Uploading...';
      case 'processing':
        return 'Processing...';
      case 'ready':
        return 'Done';
      case 'failed':
        return 'Failed';
      default:
        return status;
    }
  }

  // Event listeners
  cleanupFns.push(
    on(fileInput, 'change', (e: Event) => {
      const target = e.target as HTMLInputElement;
      if (target.files) {
        handleFiles(target.files);
        target.value = '';
      }
    }),
  );

  // Drag and drop
  function preventDefault(e: Event): void {
    e.preventDefault();
    e.stopPropagation();
  }

  cleanupFns.push(on(dropzone, 'dragover', preventDefault));
  cleanupFns.push(on(dropzone, 'dragenter', preventDefault));

  cleanupFns.push(
    on(dropzone, 'drop', (e: Event) => {
      preventDefault(e);
      const dragEvent = e as DragEvent;
      if (dragEvent.dataTransfer?.files) {
        handleFiles(dragEvent.dataTransfer.files);
      }
    }),
  );

  // Click on dropzone triggers file input
  cleanupFns.push(
    on(dropzone, 'click', () => {
      fileInput.click();
    }),
  );

  // Keyboard accessibility
  cleanupFns.push(
    on(dropzone, 'keydown', ((e: Event) => {
      const ke = e as KeyboardEvent;
      if (ke.key === 'Enter' || ke.key === ' ') {
        ke.preventDefault();
        fileInput.click();
      }
    }) as (e: Event) => void),
  );

  // Destroy
  return {
    destroy: () => {
      destroyed = true;
      for (const fn of cleanupFns) fn();
      queue.cleanup();
      removeElement(container);
    },
    getContainer: () => container,
    renderQueue,
  } as unknown as HTMLElement;
}
