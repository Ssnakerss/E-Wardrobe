# E-Wardrobe Web Client

Pure TypeScript + HTML + CSS single-page application for managing a wardrobe.

## Prerequisites

- Node.js >= 24.0.0
- npm >= 11.0.0

## Setup

```bash
npm ci
```

## Scripts

| Command                    | Description                                             |
| -------------------------- | ------------------------------------------------------- |
| `npm run dev`              | Start Vite dev server (proxy `/api` → `localhost:8080`) |
| `npm run build`            | Type-check + production build to `dist/`                |
| `npm run preview`          | Preview production build locally                        |
| `npm run typecheck`        | Run TypeScript type checking                            |
| `npm run lint`             | Run ESLint                                              |
| `npm run format:check`     | Check Prettier formatting                               |
| `npm run test`             | Run Vitest unit tests                                   |
| `npm run test:coverage`    | Run tests with coverage report                          |
| `npm run test:e2e`         | Run Playwright e2e tests                                |
| `npm run openapi:generate` | Generate TypeScript types from OpenAPI spec             |
| `npm run openapi:check`    | Check for OpenAPI spec changes                          |
| `npm run check`            | Run typecheck + lint + format + test                    |

## Project Structure
 
```
web/
├── src/
│   ├── api/            # API client, auth store, error handling
│   ├── app/            # Router, lifecycle bootstrap, types
│   ├── components/     # Reusable UI components
│   ├── dom/            # DOM helpers (element, on, clearElement)
│   ├── domain/         # Domain logic (item, category-tree)
│   ├── styles/         # Global CSS (tokens, layout, components)
│   ├── views/          # Page views (login, items, categories, etc.)
│   ├── main.ts         # Entry point
│   ├── sw.ts           # Service Worker
│   └── vite-env.d.ts   # Vite type declarations
├── e2e/                # Playwright e2e tests
├── public/             # Static assets (favicon)
├── index.html          # HTML entry point (CSP meta tag)
├── vite.config.ts      # Vite configuration
├── tsconfig.json       # TypeScript configuration
├── vitest.config.ts    # Vitest configuration
├── playwright.config.ts # Playwright configuration
└── eslint.config.js    # ESLint configuration
```

## Architecture

### Routing

Client-side routing via `app/router.ts` with hash-based navigation. Views implement the `View` interface (`mount`, `destroy`).

### API Client

`api/client.ts` wraps `fetch` with authentication (cookie-based), CSRF token handling, and error normalization (`normalizeApiError` → `ApiError` with `kind`, `httpStatus`, `message`).

### State Management

Minimal state: `auth-store.ts` holds the current user session. No global store — views manage their own state.

### Components

Custom DOM-based components (no framework). Each component exports a factory function returning an `HTMLElement`.

### Service Worker

`sw.ts` provides offline support:

- **Static assets** (JS, CSS, images): CacheFirst with background update
- **API requests**: NetworkFirst with cache fallback
- **Navigation**: NetworkFirst, falls back to `index.html` for SPA routing

Cache versioning via `CACHE_VERSION` — changes trigger cache cleanup on activate.

### Content Security Policy

`index.html` includes a CSP meta tag:

- `script-src 'self'` — only own scripts
- `style-src 'self' 'unsafe-inline'` — allows inline styles (used by the app)
- `img-src 'self' data: blob:` — allows own images, data URIs, and blob URLs (photo previews)
- `connect-src 'self'` — only same-origin API calls
- `object-src 'none'` — no plugins

## Building for Production

```bash
npm run build
```

Output: `dist/` directory with hashed assets.

### Embedding in Go

The backend embeds `dist/` via `go:embed`:

```bash
# From project root
cd web && npm run build
cd ../backend && go build ./...
```

The `webui.StaticHandler()` serves static files with SPA fallback.
