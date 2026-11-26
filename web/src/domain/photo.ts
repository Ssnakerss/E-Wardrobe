/**
 * Photo domain: upload queue with validation, concurrency control, and retry.
 */

/**
 * Possible statuses for an upload file.
 */
export type UploadStatus = 'queued' | 'uploading' | 'processing' | 'ready' | 'failed';

/**
 * Preview info for an uploaded image file.
 */
export interface ImagePreview {
  url: string;
  width: number;
  height: number;
}

/**
 * A single upload file in the queue.
 */
export interface UploadFile {
  id: string;
  file: File;
  status: UploadStatus;
  itemId: string;
  sortOrder: number;
  preview: ImagePreview;
  error: string | null;
}

/**
 * Options for the upload queue.
 */
export interface UploadQueueOptions {
  maxConcurrency?: number;
  maxSizeBytes?: number;
}

const DEFAULT_MAX_SIZE = 12 * 1024 * 1024; // 12 MiB
const JPEG_TYPE = 'image/jpeg';

/**
 * Generate a unique ID for upload files.
 */
let _idCounter = 0;

function generateId(): string {
  return `upload-${Date.now()}-${++_idCounter}`;
}

/**
 * Validate that a file is a JPEG image.
 */
export function isJpegFile(file: File): boolean {
  return file.type === JPEG_TYPE;
}

/**
 * Validate that a file is within the size limit.
 */
export function isWithinSizeLimit(file: File, maxSizeBytes: number): boolean {
  return file.size <= maxSizeBytes;
}

/**
 * Upload queue with concurrency control.
 * Processes files in enqueue order, up to maxConcurrency simultaneous uploads.
 */
export class PhotoUploadQueue {
  private _files: UploadFile[] = [];
  private _sortCounter = 0;
  private _maxConcurrency: number;
  private _maxSizeBytes: number;
  private _uploading = false;

  constructor(options: UploadQueueOptions = {}) {
    this._maxConcurrency = options.maxConcurrency ?? 3;
    this._maxSizeBytes = options.maxSizeBytes ?? DEFAULT_MAX_SIZE;
  }

  /** Current files in the queue. */
  get files(): ReadonlyArray<UploadFile> {
    return this._files;
  }

  /** Whether the queue is currently uploading. */
  get status(): UploadStatus | 'idle' {
    if (this._uploading) return 'uploading';
    if (this._files.some((f) => f.status === 'uploading')) return 'uploading';
    if (this._files.some((f) => f.status === 'queued')) return 'queued';
    // No active work — queue is idle (files may be failed/ready from past runs)
    return 'idle';
  }

  /**
   * Enqueue a file for upload.
   * Returns null if validation fails.
   */
  enqueue(file: File, itemId: string): UploadFile | null {
    // Validate JPEG
    if (!isJpegFile(file)) {
      return null;
    }

    // Validate size
    if (!isWithinSizeLimit(file, this._maxSizeBytes)) {
      return null;
    }

    // Check for duplicates (same file object)
    if (this._files.some((uf) => uf.file === file)) {
      return null;
    }

    const sortOrder = this._sortCounter++;
    const id = generateId();

    const uploadFile: UploadFile = {
      id,
      file,
      status: 'queued',
      itemId,
      sortOrder,
      preview: { url: '', width: 0, height: 0 },
      error: null,
    };

    this._files.push(uploadFile);
    return uploadFile;
  }

  /**
   * Set the preview URL for an upload file.
   * Called by the component layer after enqueue.
   */
  setPreview(fileId: string, url: string, width: number, height: number): void {
    const uploadFile = this._files.find((f) => f.id === fileId);
    if (!uploadFile) return;
    uploadFile.preview = { url, width, height };
  }

  /**
   * Start uploading all queued files.
   * Processes in order, up to maxConcurrency simultaneously.
   */
  async upload(
    itemId: string,
    uploadFn: (fileId: string, file: File, sortOrder: number) => Promise<Response>,
  ): Promise<void> {
    const queued = this._files.filter((f) => f.status === 'queued');
    if (queued.length === 0) return;

    this._uploading = true;

    // Process with concurrency limit
    const running: Promise<void>[] = [];

    for (const uploadFile of queued) {
      const promise = this._processOne(uploadFile, itemId, uploadFn);
      running.push(promise);

      if (running.length >= this._maxConcurrency) {
        await Promise.race(running);
        // Wait for all running promises to settle, then remove completed ones
        const settled = await Promise.allSettled(running);
        for (let i = running.length - 1; i >= 0; i--) {
          if (settled[i]?.status === 'fulfilled') {
            running.splice(i, 1);
          }
        }
      }
    }

    await Promise.all(running);
    this._uploading = false;
  }

  private async _processOne(
    uploadFile: UploadFile,
    _itemId: string,
    uploadFn: (fileId: string, file: File, sortOrder: number) => Promise<Response>,
  ): Promise<void> {
    uploadFile.status = 'uploading';

    try {
      const response = await uploadFn(uploadFile.id, uploadFile.file, uploadFile.sortOrder);

      if (!response.ok) {
        uploadFile.status = 'failed';
        uploadFile.error = `Upload failed: ${response.status}`;
        return;
      }

      uploadFile.status = 'processing';

      // Wait a tick for processing status to be visible
      await new Promise((resolve) => setTimeout(resolve, 0));

      uploadFile.status = 'ready';
    } catch (error) {
      uploadFile.status = 'failed';
      uploadFile.error = error instanceof Error ? error.message : 'Upload failed';
    }
  }

  /**
   * Retry a failed upload.
   */
  async retry(
    fileId: string,
    itemId: string,
    uploadFn: (fileId: string, file: File, sortOrder: number) => Promise<Response>,
  ): Promise<void> {
    const uploadFile = this._files.find((f) => f.id === fileId);
    if (!uploadFile || uploadFile.status !== 'failed') return;

    uploadFile.status = 'queued';
    await this.upload(itemId, uploadFn);
  }

  /**
   * Remove a file from the queue.
   */
  dequeue(fileId: string): void {
    const index = this._files.findIndex((f) => f.id === fileId);
    if (index === -1) return;
    this._files.splice(index, 1);
  }

  /**
   * Get files filtered by status.
   */
  getByStatus(status: UploadStatus): ReadonlyArray<UploadFile> {
    return this._files.filter((f) => f.status === status);
  }

  /**
   * Clear all files.
   */
  cleanup(): void {
    this._files = [];
    this._sortCounter = 0;
    this._uploading = false;
  }
}
