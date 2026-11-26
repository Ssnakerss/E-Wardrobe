/**
 * Unit tests for photo upload queue logic.
 */

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { PhotoUploadQueue } from './photo.js';

// Mock fetch
(globalThis as Record<string, unknown>).fetch = vi.fn();

function mockResponse(body: unknown, status = 200): Response {
  return {
    ok: status >= 200 && status < 300,
    status,
    statusText: status === 200 ? 'OK' : 'Error',
    json: async () => body,
    headers: new Headers([['Content-Type', 'application/json']]),
    redirected: false,
    type: 'basic' as ResponseType,
    url: '',
    clone: () => mockResponse(body, status),
    body: null,
    bodyUsed: false,
    arrayBuffer: async () => new ArrayBuffer(0),
    blob: async () => new Blob(),
    formData: async () => new FormData(),
    text: async () => JSON.stringify(body),
  } as unknown as Response;
}

function createFile(name = 'test.jpg', size = 1024, type = 'image/jpeg'): File {
  return new File([new ArrayBuffer(size)], name, { type });
}

describe('PhotoUploadQueue', () => {
  beforeEach(() => {
    (globalThis.fetch as ReturnType<typeof vi.fn>).mockClear();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('starts with empty queue', () => {
    const queue = new PhotoUploadQueue({ maxConcurrency: 2 });
    expect(queue.files).toEqual([]);
    expect(queue.status).toBe('idle');
  });

  it('validates JPEG type on enqueue', () => {
    const queue = new PhotoUploadQueue({ maxConcurrency: 2 });
    const pngFile = new File([new ArrayBuffer(100)], 'test.png', { type: 'image/png' });
    const result = queue.enqueue(pngFile, 'item-1');
    expect(result).toBeNull();
    expect(queue.files).toEqual([]);
  });

  it('validates file size on enqueue', () => {
    const queue = new PhotoUploadQueue({ maxConcurrency: 2, maxSizeBytes: 1024 });
    const bigFile = new File([new ArrayBuffer(2048)], 'big.jpg', { type: 'image/jpeg' });
    const result = queue.enqueue(bigFile, 'item-1');
    expect(result).toBeNull();
    expect(queue.files).toEqual([]);
  });

  it('enqueues valid file with correct status', () => {
    const queue = new PhotoUploadQueue({ maxConcurrency: 2 });
    const file = createFile();
    const uploadFile = queue.enqueue(file, 'item-1')!;

    expect(uploadFile.status).toBe('queued');
    expect(uploadFile.file).toBe(file);
    expect(uploadFile.itemId).toBe('item-1');
    expect(uploadFile.sortOrder).toBe(0);
    expect(uploadFile.preview).toEqual({ url: '', width: 0, height: 0 });
  });

  it('increments sort_order for each enqueue', () => {
    const queue = new PhotoUploadQueue({ maxConcurrency: 2 });
    const file1 = createFile();
    const file2 = createFile();
    const file3 = createFile();

    const uf1 = queue.enqueue(file1, 'item-1')!;
    const uf2 = queue.enqueue(file2, 'item-1')!;
    const uf3 = queue.enqueue(file3, 'item-1')!;

    expect(uf1.sortOrder).toBe(0);
    expect(uf2.sortOrder).toBe(1);
    expect(uf3.sortOrder).toBe(2);
  });

  it('rejects duplicate files by identity', () => {
    const queue = new PhotoUploadQueue({ maxConcurrency: 2 });
    const file = createFile();
    const first = queue.enqueue(file, 'item-1');
    const second = queue.enqueue(file, 'item-1');

    expect(first).not.toBeNull();
    expect(second).toBeNull();
    expect(queue.files).toHaveLength(1);
  });

  it('processes files in order up to maxConcurrency', async () => {
    vi.useFakeTimers();
    const uploadOrder: string[] = [];
    const resolvePromises: (() => void)[] = [];

    (globalThis.fetch as ReturnType<typeof vi.fn>).mockImplementation(async () => {
      return new Promise<Response>((resolve) => {
        const promiseResolve = () => {
          resolve(mockResponse({ id: 'photo-1' }));
        };
        resolvePromises.push(promiseResolve);
      });
    });

    const queue = new PhotoUploadQueue({ maxConcurrency: 2 });
    queue.enqueue(createFile('a.jpg'), 'item-1');
    queue.enqueue(createFile('b.jpg'), 'item-1');
    queue.enqueue(createFile('c.jpg'), 'item-1');

    void queue.upload('item-1', (_fileId, file, _sortOrder) => {
      uploadOrder.push(file.name);
      return new Promise<Response>((resolve) => {
        const promiseResolve = () => {
          resolve(mockResponse({ id: `photo-${file.name}` }));
        };
        resolvePromises.push(promiseResolve);
      });
    });

    // upload() starts processing synchronously — first 2 files pushed to order
    expect(queue.status).toBe('uploading');
    expect(uploadOrder).toHaveLength(2); // maxConcurrency = 2, processed synchronously

    // Resolve first upload
    resolvePromises.shift()!();
    await vi.advanceTimersByTimeAsync(0);
    expect(uploadOrder).toContain('a.jpg');

    // Resolve second upload — third should start
    resolvePromises.shift()!();
    await vi.advanceTimersByTimeAsync(0);
    expect(uploadOrder).toContain('b.jpg');

    // Resolve third
    resolvePromises.shift()!();
    await vi.advanceTimersByTimeAsync(0);
    expect(uploadOrder).toContain('c.jpg');
    expect(queue.status).toBe('idle');
    vi.useRealTimers();
  });

  it('marks file as failed on upload error', async () => {
    (globalThis.fetch as ReturnType<typeof vi.fn>).mockRejectedValueOnce(new Error('network'));

    const queue = new PhotoUploadQueue({ maxConcurrency: 1 });
    queue.enqueue(createFile(), 'item-1');

    await queue.upload('item-1', async () => {
      throw new Error('upload failed');
    });

    expect(queue.files[0]!.status).toBe('failed');
    expect(queue.status).toBe('idle');
  });

  it('retries failed uploads on retry', async () => {
    let callCount = 0;
    (globalThis.fetch as ReturnType<typeof vi.fn>).mockImplementation(async () => {
      callCount++;
      if (callCount === 1) {
        throw new Error('first fail');
      }
      return mockResponse({ id: 'photo-1' });
    });

    const queue = new PhotoUploadQueue({ maxConcurrency: 1 });
    queue.enqueue(createFile(), 'item-1');

    await queue.upload('item-1', async () => {
      throw new Error('fail');
    });

    expect(queue.files[0]!.status).toBe('failed');

    // Retry
    await queue.retry(queue.files[0]!.id, 'item-1', async () => {
      return mockResponse({ id: 'photo-1-retried' });
    });

    expect(queue.files[0]!.status).toBe('ready');
  });

  it('removes file from queue on dequeue', () => {
    const queue = new PhotoUploadQueue({ maxConcurrency: 2 });
    const file = createFile();
    const result = queue.enqueue(file, 'item-1');
    expect(result).not.toBeNull();
    expect(queue.files).toHaveLength(1);

    queue.dequeue(queue.files[0]!.id);
    expect(queue.files).toHaveLength(0);
  });

  it('clears all files on cleanup', () => {
    const queue = new PhotoUploadQueue({ maxConcurrency: 2 });
    queue.enqueue(createFile(), 'item-1');
    queue.enqueue(createFile(), 'item-1');
    expect(queue.files).toHaveLength(2);

    queue.cleanup();
    expect(queue.files).toHaveLength(0);
  });

  it('filters files by status', () => {
    const queue = new PhotoUploadQueue({ maxConcurrency: 2 });
    const uf1 = queue.enqueue(createFile(), 'item-1');
    if (uf1) uf1.status = 'ready';
    const uf2 = queue.enqueue(createFile(), 'item-1');
    if (uf2) uf2.status = 'queued';

    const ready = queue.getByStatus('ready');
    const queued = queue.getByStatus('queued');

    expect(ready).toHaveLength(1);
    expect(queued).toHaveLength(1);
  });

  it('generates unique IDs for each upload file', () => {
    const queue = new PhotoUploadQueue({ maxConcurrency: 2 });
    const uf1 = queue.enqueue(createFile(), 'item-1');
    const uf2 = queue.enqueue(createFile('other.jpg'), 'item-1');

    expect(uf1!.id).not.toBe(uf2!.id);
  });

  it('sets preview after enqueue', () => {
    const queue = new PhotoUploadQueue({ maxConcurrency: 2 });
    const uf = queue.enqueue(createFile(), 'item-1')!;

    expect(uf.preview.url).toBe('');
    queue.setPreview(uf.id, 'blob:test-123', 800, 600);

    expect(queue.files[0]!.preview).toEqual({ url: 'blob:test-123', width: 800, height: 600 });
  });

  it('returns correct status for mixed states', () => {
    const queue = new PhotoUploadQueue({ maxConcurrency: 2 });
    const uf1 = queue.enqueue(createFile(), 'item-1');
    const uf2 = queue.enqueue(createFile(), 'item-1');

    if (uf1) uf1.status = 'uploading';
    if (uf2) uf2.status = 'queued';

    expect(queue.status).toBe('uploading');
  });
});
