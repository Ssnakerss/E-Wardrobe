/**
 * Service Worker — caches static assets with CacheFirst,
 * API requests with NetworkFirst, and index.html with NetworkFirst.
 */

declare const self: ServiceWorkerGlobalScope;

const CACHE_VERSION = 'v1';
const CACHE_NAME = `wardrobe-${CACHE_VERSION}`;

// Assets to precache on install
const PRECACHE_URLS = ['/', '/index.html'];

// Install: precache static assets
self.addEventListener('install', (event: ExtendableEvent) => {
  event.waitUntil(caches.open(CACHE_NAME).then((cache) => cache.addAll(PRECACHE_URLS)));
  // Activate immediately
  self.skipWaiting();
});

// Activate: clean old caches
self.addEventListener('activate', (event: ExtendableEvent) => {
  event.waitUntil(
    caches
      .keys()
      .then((keys) =>
        Promise.all(keys.filter((key) => key !== CACHE_NAME).map((key) => caches.delete(key))),
      ),
  );
  // Take control of all clients immediately
  void self.clients.claim();
});

// Fetch: route-based strategy
self.addEventListener('fetch', (event: FetchEvent) => {
  const { request } = event;

  // Only handle GET requests
  if (request.method !== 'GET') return;

  const url = new URL(request.url);

  // API requests: NetworkFirst
  if (url.pathname.startsWith('/api/')) {
    event.respondWith(networkFirst(request));
    return;
  }

  // index.html: NetworkFirst
  if (request.destination === 'document') {
    event.respondWith(networkFirst(request));
    return;
  }

  // Static assets (JS, CSS, images, fonts): CacheFirst
  event.respondWith(cacheFirst(request));
});

/**
 * CacheFirst strategy: try cache, then network.
 */
async function cacheFirst(request: Request): Promise<Response> {
  const cache = await caches.open(CACHE_NAME);
  const cached = await cache.match(request);

  if (cached) {
    // Serve from cache, but also update in background
    const updatePromise = cache.add(request).catch(() => {
      // Ignore failures (e.g., network down)
    });
    self.clients.matchAll().then((clients: unknown[]) => {
      if (clients.length > 0) {
        (clients[0] as { postMessage: (msg: unknown) => void }).postMessage({
          type: 'CACHE_UPDATED',
        });
      }
    });
    void updatePromise;
    return cached;
  }

  try {
    const networkResponse = await fetch(request);
    if (networkResponse.ok) {
      await cache.put(request, networkResponse.clone());
    }
    return networkResponse;
  } catch {
    // Return a generic offline response for images, 404 for others
    if (request.destination === 'image') {
      return new Response('', { status: 404, statusText: 'Offline' });
    }
    return new Response('Offline', { status: 503, statusText: 'Service Unavailable' });
  }
}

/**
 * NetworkFirst strategy: try network, then cache fallback.
 */
async function networkFirst(request: Request): Promise<Response> {
  try {
    const networkResponse = await fetch(request);
    if (networkResponse.ok) {
      // Cache the response for future use
      const cache = await caches.open(CACHE_NAME);
      await cache.put(request, networkResponse.clone());
    }
    return networkResponse;
  } catch {
    // Fallback to cache
    const cache = await caches.open(CACHE_NAME);
    const cached = await cache.match(request);
    if (cached) {
      return cached;
    }
    // For navigation requests, return index.html for SPA routing
    if (request.destination === 'document') {
      const indexHtml = await cache.match('/index.html');
      if (indexHtml) return indexHtml;
    }
    return new Response('Offline', { status: 503, statusText: 'Service Unavailable' });
  }
}

// Message handler for cache management
self.addEventListener('message', (event: MessageEvent) => {
  const data = event.data as { type?: string } | undefined;

  if (data?.type === 'SKIP_WAITING') {
    self.skipWaiting();
  }

  if (data?.type === 'CLEAR_CACHE') {
    caches.delete(CACHE_NAME);
  }
});
