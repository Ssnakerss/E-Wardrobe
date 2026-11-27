/// <reference types="vite/client" />

// Service Worker global types
interface ServiceWorkerGlobalScope {
  skipWaiting(): void;
  clients: Clients;
  addEventListener(type: 'install' | 'activate', listener: (event: ExtendableEvent) => void): void;
  addEventListener(type: 'fetch', listener: (event: FetchEvent) => void): void;
  addEventListener(type: 'message', listener: (event: MessageEvent) => void): void;
}

interface Clients {
  matchAll(): Promise<unknown[]>;
  claim(): Promise<void>;
}

interface ExtendableEvent {
  waitUntil(fn: Promise<unknown>): void;
}

interface FetchEvent {
  request: Request;
  respondWith(response: Promise<Response>): void;
}
