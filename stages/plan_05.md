Вот промт для Этапа 5:

---

**Продолжение реализации веб-клиента E-Wardrobe (Этап 5 из 8) — UX, Accessibility, Адаптивная стабилизация**

## Контекст
- **Проект:** E-Wardrobe — управление гардеробом
- **Бэкенд:** Go backend, API по `docs/openapi.yaml`
- **Фронтенд:** `web/` — чистый TypeScript + HTML + CSS, Vite, без UI-фреймворков
- **План:** `docs/web-implementation-plan.md`

## Текущий прогресс (5 этапов завершены)

### Завершённые этапы:
- **Этап 0** ✅ — Scaffold: Vite + TypeScript strict + ESLint + Prettier + Vitest + Playwright + OpenAPI types
- **Этап 1** ✅ — App shell, router, API client, auth (login/register/recovery)
- **Этап 2** ✅ — Категории: дерево, CRUD, навигация, 19 unit tests
- **Этап 3** ✅ — Вещи: список (карточки/таблица), фильтры, поиск, детали, формы, архив, delete, conflict dialog, 15 unit tests
- **Этап 4** ✅ — Фотографии: upload (multipart, queue), preview, gallery, lightbox, reorder, delete

## Коммиты:
```
<commit-этап-4> feat(web): photos upload, gallery, reorder, lightbox
1569458 feat(web): items list, details, forms, archive, delete
f26d104 feat(web): implement categories management with tree navigation
a91a83f feat(web): app shell, router, API client, auth views
8fc7884 feat(web): scaffold Vite project with TypeScript, ESLint, Vitest, Playwright
```

## Команда для проверки:
```
cd web && npm run check   # typecheck + lint + format + test
npm run build              # production build
```

---

# Этап 5: UX, Accessibility, Адаптивная стабилизация

**Из плана (раздел 13 и 15, Этап 5):**

### Задачи:

#### 1. Проверить и дополнить все loading/empty/error states
- **План:** "Проверить все loading/empty/error states"
- Каждая view должна иметь корректное отображение для:
  - Загрузки (loading-state с сообщением)
  - Пустого результата (empty-state с CTA)
  - Ошибки сети (error-state с retry)
  - Ошибки 404 (ресурс не найден — не как ошибка авторизации)
- **Что проверить:**
  - `items-view.ts`: пустой результат, ошибка сети, 404 чужой категории
  - `item-details-view.ts`: загрузка, 404, ошибка сети, conflict
  - `categories-view.ts`: загрузка, пустое дерево, ошибка
  - `login-view.ts`: rate limiting (429), сетевая ошибка
  - `item-form-view.ts`: ошибка валидации, save fail

#### 2. Довести responsive table/cards
- **План:** "Довести responsive table/cards"
- Проверить breakpoints:
  - Узкий телефон (~360px)
  - Широкий телефон (~420px)
  - Планшет (~768px)
  - Стандартный desktop (~1024px)
  - Широкий desktop (~1440px)
- **Конкретно:**
  - `.items-table` на мобильных — horizontal scroll, не ломать layout
  - `.item-card` grid — 1 колонка на узком экране
  - `.view-header` — vertical layout на mobile
  - `.dialog` — width: 95% на mobile, padding
  - `.item-filters` — wrap на mobile
  - `.quick-actions` — wrap на mobile

#### 3. Проверить длинные названия, описания и места хранения
- **План:** "Проверить длинные названия, описания и места хранения"
  - `.item-card__title` — truncate с `...` (`overflow: hidden; text-overflow: ellipsis; white-space: nowrap`)
  - `.item-details__value` — wrap длинного текста (`word-wrap: break-word`)
  - `.error-state__message`, `.toast__message` — wrap
  - `.empty-state__description` — wrap

#### 4. Настроить Service Worker (кэширование статики)
- **План:** "Service Worker — кэширование статики"
- `web/src/sw.ts` — Service Worker с кэшированием:
  - Статические файлы (JS, CSS, images) — cache-first с stale-while-revalidate
  - API запросы (`/api/*`) — network-first, fallback на cache
  - `index.html` — network-first
  - Кэш-стратегия: CacheFirst для статики, NetworkFirst для API
  - Регистрация SW в `main.ts`
  - Версионирование кэша через `CACHE_VERSION`

#### 5. Настроить CSP (Content Security Policy)
- **План:** "CSP"
- Добавить мета-тег CSP в `index.html`:
  ```html
  <meta http-equiv="Content-Security-Policy" content="default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: blob:; connect-src 'self'; object-src 'none'; base-uri 'self'; form-action 'self'">
  ```
- Убедиться, что inline-стили (если есть) добавлены в `style-src 'unsafe-inline'`
- Проверить, что все ресурсы загружаются по CSP

#### 6. Настроить HTTPS в dev
- **План:** "HTTPS в dev"
- Настроить Vite dev server с HTTPS (self-signed cert)
- `vite.config.ts`: `server: { https: true }`
- Для тестов — использовать `https: { selfSigned: true }`

#### 7. Интегрировать go:embed (базовая версия)
- **План:** "go:embed" — встроить собранную статику в Go binary
- В `backend/` добавить `go:embed web/dist`
- HTTP handler для раздачи static files
- SPA fallback: все не-API запросы → `index.html`
- Проверить, что `go build` проходит

#### 8. Настроить CI hardening
- **План:** "CI hardening"
- Добавить `.github/workflows/web-ci.yml`:
  - `npm ci`
  - `npm run typecheck`
  - `npm run lint`
  - `npm run format:check`
  - `npm run test`
  - `npm run build`
  - Cache node_modules
  - Fail on any error

#### 9. Написать документацию
- **План:** "документация"
- `web/README.md`:
  - Установка: `npm ci`
  - Dev: `npm run dev`
  - Build: `npm run build`
  - Tests: `npm run test`, `npm run test:coverage`, `npm run test:e2e`
  - Lint: `npm run lint`, `npm run format:check`
  - OpenAPI: `npm run openapi:generate`
  - Структура проекта
  - Архитектурные решения

---

### Файлы для создания/изменения:

1. **`web/src/sw.ts`** — Service Worker с кэш-стратегиями
2. **`web/src/main.ts`** — регистрация SW
3. **`web/index.html`** — CSP meta tag
4. **`web/src/styles/index.css`** — truncate для длинных текстов, улучшенные breakpoints
5. **`web/src/views/*.ts`** — проверить/добавить loading/empty/error states
6. **`backend/handler.go`** (или аналог) — `go:embed web/dist`, SPA fallback
7. **`.github/workflows/web-ci.yml`** — CI pipeline
8. **`web/README.md`** — документация

---

### Ключевые требования:
- TypeScript strict, ESLint проходит
- Все существующие тесты проходят
- `npm run build` создаёт рабочую сборку
- CSP не ломает загрузку ресурсов
- Service Worker не блокирует основной поток
- go:embed не ломает `go build`
- CI pipeline проходит за < 5 минут
- Документация покрывает все npm scripts

### Что НЕ делать:
- Не менять API contract
- Не добавлять новые зависимости (SW — нативный)
- Не менять существующую логику views
- Не добавлять UI-фреймворки

---

**Начни с `npm run check` и `npm run build`, чтобы убедиться что текущий код валиден. Затем реализуй Этап 5 поэтапно. После завершения — commit & push.**