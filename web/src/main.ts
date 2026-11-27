import { bootstrap } from './app/lifecycle-bootstrap.js';

/**
 * Register Service Worker for offline support.
 */
async function registerServiceWorker(): Promise<void> {
  if ('serviceWorker' in navigator) {
    try {
      // Only register in production (HTTPS or localhost)
      const isSecure = location.protocol === 'https:' || location.hostname === 'localhost';
      if (!isSecure) return;

      const swUrl = '/sw.js';
      const registration = await navigator.serviceWorker.register(swUrl, { scope: '/' });
      console.log('SW registered:', registration.scope);

      // Check for updates every hour
      setInterval(async () => {
        try {
          const updated = await registration.update();
          if (updated.installing) {
            console.log('SW update available');
          }
        } catch {
          // Update check failed silently
        }
      }, 3600000);
    } catch (error) {
      console.warn('Failed to register SW:', error);
    }
  }
}

bootstrap().catch((error) => {
  console.error('Failed to bootstrap app:', error);
  const app = document.getElementById('app');
  if (app) {
    app.innerHTML = `
      <div class="error-state" role="alert">
        <h2>Application Error</h2>
        <p>Failed to initialize. Check the console for details.</p>
      </div>
    `;
  }
});

// Register SW after bootstrap starts
registerServiceWorker();
