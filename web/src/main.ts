import { bootstrap } from './app/lifecycle-bootstrap.js';

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
