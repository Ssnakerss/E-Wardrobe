# Stage 0 Plan: OpenAPI Contracts + Vite Scaffold

**Дата:** 2026-11-25  
**Этап:** 0 из 8  
**Оценка:** 2–3 дня

---

## 1. Задача

Создать рабочую инфраструктуру веб-проекта:
- Проверить OpenAPI-контракт на соответствие требованиям плана (раздел 3)
- Сгенерировать TypeScript-типы из OpenAPI
- Настроить Vite + TypeScript strict + ESLint + Vitest + Playwright
- Настроить npm scripts, lockfile, фиксированную Node.js LTS
- Настроить dev proxy на бэкенд

---

## 2. OpenAPI Gap Analysis (план раздел 3.1)

### Проверено — все endpoints присутствуют:

| Endpoint | Статус | Примечание |
|----------|--------|------------|
| `POST /api/v1/auth/invitations/validate` | ✅ Есть | |
| `POST /api/v1/auth/recovery/validate` | ✅ Есть | |
| `POST /api/v1/auth/recovery/complete` | ✅ Есть | |
| `PUT /api/v1/items/{itemId}/photos/order` | ✅ Есть | |
| `POST /api/v1/auth/register` | ✅ Есть | |
| `POST /api/v1/auth/login` | ✅ Есть | |
| `POST /api/v1/auth/logout` | ✅ Есть | |
| `GET /api/v1/auth/csrf` | ✅ Есть | |
| `GET /api/v1/auth/me` | ✅ Есть | |
| `GET/POST /api/v1/categories` | ✅ Есть | |
| `PATCH/DELETE /api/v1/categories/{id}` | ✅ Есть | |
| `GET/POST /api/v1/items` | ✅ Есть | |
| `GET/PATCH/DELETE /api/v1/items/{id}` | ✅ Есть | |
| `POST /api/v1/items/{id}/archive` | ✅ Есть | |
| `POST /api/v1/items/{id}/photos` | ✅ Есть | |
| `DELETE /api/v1/photos/{id}` | ✅ Есть | |

### Требуют уточнения (план раздел 3.1):

1. **`409 version_conflict`** — в OpenAPI `409` ссылается на `$ref: '#/components/responses/Error'`, но план требует уточнения:
   - Должен ли ответ содержать `current` snapshot?
   - Должен ли ответ содержать `current_version`?
   - **Решение:** OpenAPI уже содержит `SyncMutationResult.current` для sync-конфликтов. Для REST API нужно добавить явный schema для 409 response с `current` и `current_version`.

2. **Коды ошибок CSRF, validation, invitation, recovery, upload** — `ErrorResponse` имеет generic `code: string`. План требует конкретных кодов.
   - **Решение:** Добавить enum в описание или documentation. Frontend будет мапить по HTTP status + code string.

3. **Ограничения длины текстовых полей** — есть в schema (name: 1-200, password: 12-128), но не документированы в responses.
   - **Решение:** Добавить `maxLength` в response error details.

4. **Ограничения `purchase_year`** — нет явных min/max в schema.
   - **Решение:** Добавить `minimum: 1900` и `maximum: 2100` к `purchase_year`.

5. **Правила очистки nullable-полей через JSON `null`** — schema использует `nullable: true`, но не документировано поведение PATCH.
   - **Решение:** Документация в плане достаточна. Frontend реализует отправку `null` для очистки.

6. **`current snapshot` при конфликте** — в REST PATCH/DELETE 409 response не содержит current entity.
   - **Решение:** Для Stage 0 принять текущий контракт. На Этапе 3 при реализации conflict dialog добавить GET для загрузки current entity при 409.

### Итог по OpenAPI:
Контракт **достаточен для начала разработки**. Критические пробелы:
- `purchase_year` без ограничений (низкий приоритет, валидация на frontend)
- 409 response без current snapshot (решается на Этапе 3 через отдельный GET)

**Рекомендация:** Начать разработку, мелкие правки OpenAPI добавить параллельно.

---

## 3. Структура проекта

```
web/
├── public/
│   ├── favicon.svg
│   └── robots.txt
├── src/
│   ├── main.ts                    # Entry point
│   ├── app/
│   │   ├── config.ts              # VITE_ env vars (no secrets)
│   │   └── lifecycle.ts           # View lifecycle types
│   ├── tests/
│   │   └── setup.ts               # Vitest globals setup
│   └── styles/
│       └── index.css              # Empty placeholder
├── e2e/
│   └── _placeholder.spec.ts       # Playwright placeholder
├── index.html
├── package.json
├── package-lock.json
├── tsconfig.json
├── tsconfig.node.json
├── vite.config.ts
├── vitest.config.ts
├── playwright.config.ts
├── eslint.config.js
├── .node-version
├── .nvmrc
├── .gitignore
└── .eslintrc.cjs                  # Legacy for @typescript-eslint/parser compat
```

---

## 4. Конфигурации

### 4.1. package.json

```json
{
  "name": "wardrobe-web",
  "private": true,
  "version": "0.0.0",
  "type": "module",
  "engines": {
    "node": ">=24.0.0",
    "npm": ">=11.0.0"
  },
  "scripts": {
    "dev": "vite",
    "build": "tsc --noEmit && vite build",
    "preview": "vite preview",
    "typecheck": "tsc --noEmit",
    "lint": "eslint .",
    "format:check": "prettier --check .",
    "test": "vitest run",
    "test:coverage": "vitest run --coverage",
    "test:e2e": "playwright test",
    "openapi:generate": "openapi-typescript docs/openapi.yaml -o src/api/generated/schema.d.ts",
    "openapi:check": "node scripts/check-openapi-diff.mjs",
    "check": "npm run typecheck && npm run lint && npm run format:check && npm run test"
  },
  "devDependencies": {
    "@playwright/test": "^1.54.0",
    "@typescript-eslint/eslint-plugin": "^8.39.0",
    "@typescript-eslint/parser": "^8.39.0",
    "@vitest/coverage-v8": "^3.2.4",
    "eslint": "^9.34.0",
    "openapi-typescript": "^7.8.0",
    "prettier": "^3.6.2",
    "typescript": "~5.9.0",
    "vite": "^7.1.10",
    "vitest": "^3.2.4"
  }
}
```

### 4.2. tsconfig.json

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "useDefineForClassFields": true,
    "module": "ESNext",
    "lib": ["ES2022", "DOM", "DOM.Iterable"],
    "skipLibCheck": true,
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "isolatedModules": true,
    "moduleDetection": "force",
    "noEmit": true,
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true,
    "noUncheckedIndexedAccess": true,
    "forceConsistentCasingInFileNames": true,
    "paths": {
      "@/*": ["./src/*"]
    }
  },
  "include": ["src", "e2e"]
}
```

### 4.3. vite.config.ts

```ts
import { defineConfig } from 'vite';
import { resolve } from 'path';

export default defineConfig({
  plugins: [],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: false,
  },
});
```

### 4.4. vitest.config.ts

```ts
import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/tests/setup.ts'],
    include: ['src/**/*.test.ts'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'lcov'],
      exclude: ['src/api/generated/', 'src/tests/'],
    },
  },
});
```

### 4.5. playwright.config.ts

```ts
import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: 0,
  workers: process.env.CI ? 1 : undefined,
  use: {
    baseURL: 'http://localhost:5173',
    trace: 'on-first-retry',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
  webServer: {
    command: 'npm run dev',
    url: 'http://localhost:5173',
    reuseExistingServer: !process.env.CI,
  },
});
```

### 4.6. eslint.config.js

```js
import tsParser from '@typescript-eslint/parser';
import tsPlugin from '@typescript-eslint/eslint-plugin';

export default [
  {
    files: ['src/**/*.ts'],
    languageOptions: {
      parser: tsParser,
      parserOptions: {
        ecmaVersion: 2022,
        sourceType: 'module',
      },
    },
    plugins: {
      '@typescript-eslint': tsPlugin,
    },
    rules: {
      ...tsPlugin.configs.strict.rules,
      '@typescript-eslint/no-explicit-any': 'warn',
    },
  },
];
```

### 4.7. .prettierrc

```json
{
  "semi": true,
  "trailingComma": "all",
  "singleQuote": true,
  "printWidth": 100,
  "tabWidth": 2
}
```

---

## 5. Зависимости

| Package | Version | Purpose |
|---------|---------|---------|
| vite | ^7.1.10 | Build tool |
| typescript | ~5.9.0 | Type checking |
| openapi-typescript | ^7.8.0 | Generate TS types from OpenAPI |
| @typescript-eslint/eslint-plugin | ^8.39.0 | Linting |
| @typescript-eslint/parser | ^8.39.0 | TS parser for ESLint |
| eslint | ^9.34.0 | Linting |
| prettier | ^3.6.2 | Formatting |
| vitest | ^3.2.4 | Unit testing |
| @vitest/coverage-v8 | ^3.2.4 | Coverage |
| jsdom | (via vitest) | DOM environment |
| @playwright/test | ^1.54.0 | E2E testing |

**Zero runtime dependencies.** No UI framework, no state library.

---

## 6. Критерии готовности

- [x] `npm ci` работает на чистом checkout
- [x] `npm run typecheck` проходит без ошибок
- [x] `npm run lint` проходит без ошибок
- [x] `npm run format:check` проходит
- [x] `npm run test` проходит (пустой suite)
- [x] Dev server открывается и показывает минимальную страницу
- [x] `/api` проксируется на localhost:8080
- [x] `npm run build` создаёт dist/ без ошибок
- [x] `openapi:generate` создаёт schema.d.ts

---

## 7. Порядок выполнения

1. Создать `package.json`, `.node-version`, `.nvmrc`
2. `npm init` / `npm install` — установить devDependencies
3. Создать `tsconfig.json`, `tsconfig.node.json`
4. Создать `vite.config.ts`
5. Создать `vitest.config.ts`
6. Создать `playwright.config.ts`
7. Создать `eslint.config.js`
8. Создать `.prettierrc`
9. Создать `index.html`, `src/main.ts`
10. Создать `src/tests/setup.ts`
11. Создать `public/` файлы
12. Создать `e2e/_placeholder.spec.ts`
13. Создать `scripts/check-openapi-diff.mjs`
14. Настроить `.gitignore`
15. Запустить `npm ci && npm run check` — убедиться что всё проходит
16. Запустить `npm run openapi:generate` — проверить генерацию типов
