**Продолжение реализации веб-клиента E-Wardrobe (Этап 4 из 8)**

## Контекст
- **Проект:** E-Wardrobe — управление гардеробом
- **Бэкенд:** Go backend, API по `docs/openapi.yaml`
- **Фронтенд:** `web/` — чистый TypeScript + HTML + CSS, Vite, без UI-фреймворков
- **План:** `docs/web-implementation-plan.md`

## Текущий прогресс (4 этапа завершены)

### Этап 0 ✅ — Scaffold
- Vite + TypeScript strict + ESLint + Prettier + Vitest + Playwright
- OpenAPI types генерация из `docs/openapi.yaml`
- npm scripts: dev, build, typecheck, lint, test, test:e2e, openapi:generate
- Все проверки проходят: `npm run check`

### Этап 1 ✅ — App shell, router, auth
- **Router:** History API router с lazy loading views, guards, param matching
- **API Client:** `apiGet/Post/Patch/Delete`, CSRF auto-handling, error normalization (ApiErrorKind), abort signals
- **Auth Store:** bootstrap, login, register, recovery, logout
- **Views:** login-view, register-view (с invitation), recovery-view
- **Components:** app-header, toast, loading-state, empty-state, error-state
- **DOM helpers:** element(), on(), template, focus management
- **CSS:** tokens, reset, forms, responsive layout

### Этап 2 ✅ — Категории
- **API:** `fetchCategories`, `createCategory`, `updateCategory`, `deleteCategory`, `archiveCategory`
- **Domain:** `buildCategoryTree`, `getCategoryById`, `getAllDescendantIds`, `getAvailableParents` (19 unit tests)
- **Components:** category-nav (дерево с раскрытием, rename/move/delete), category-select
- **View:** categories-view с полной реализацией

### Этап 3 ✅ — Вещи
- **API:** `fetchItems`, `fetchItem`, `createItem`, `updateItem`, `archiveItem`, `deleteItem`, `fetchStorageLocations`
- **Domain:** `buildItemPatch` (minimal patch), `parsePrice` (строковые целые), `formatPrice` (Intl), `validateItemForm` (15 unit tests)
- **Views:** items-view (поиск с debounce, фильтры, сортировка, cursor pagination, cards/table toggle), item-details-view (все поля, быстрые действия), item-form-view (16 полей, domain validation)
- **Components:** item-card, item-filters, confirm-dialog, conflict-dialog (409 version conflict)
- **CSS:** items list, cards grid, table, dialogs, responsive

## Коммиты в git:
```
1569458 feat(web): items list, details, forms, archive, delete
f26d104 feat(web): implement categories management with tree navigation
a91a83f feat(web): app shell, router, API client, auth views
8fc7884 feat(web): scaffold Vite project with TypeScript, ESLint, Vitest, Playwright
```

## Что делать дальше — Этап 4: Фотографии

**Из плана (раздел 11 и 15):**

### Задачи:
1. **API:** `uploadPhoto(itemId, file, sortOrder)`, `deletePhoto(photoId, baseVersion)`, `reorderPhotos(itemId, orderedPhotos[])`
2. **Component: photo-uploader.ts** — выбор JPEG файлов, preview через `URL.createObjectURL`, multipart upload с `FormData`, очередь с ограниченным concurrency, per-file status (queued/uploading/processing/ready/failed), retry failed uploads
3. **Component: photo-gallery.ts** — галерея с thumbnail/original, lightbox через `<dialog>`, focus trap, Escape to close, lazy loading
4. **Component: photo-order.ts** — reorder через кнопки "влево/вправо", отправка ordered ID list + versions
5. **View: item-details-view.ts** — интегрировать photo gallery, uploader, reorder в детали вещи
6. **CSS:** стили для галереи, uploader, lightbox, status badges
7. **Tests:** unit tests для upload queue logic

### Ключевые правила из плана:
- JPEG only, max 12 MiB per file
- `POST /api/v1/items/{itemId}/photos` с `multipart/form-data`
- `DELETE /api/v1/photos/{photoId}?base_version=N`
- `PUT /api/v1/items/{itemId}/photos/order` — полный ordered ID list + versions
- `URL.createObjectURL` → `URL.revokeObjectURL` на destroy view
- Ошибка фото НЕ удаляет вещь и НЕ блокирует другие файлы
- Обложка = фото с минимальным `sort_order`
- Conflict (409) при reorder — перечитать порядок, не считать успешным
- Keyboard accessible reorder (не только drag-and-drop)

### Структура существующего кода:
- API: `web/src/api/items.ts` (добавить фото endpoints)
- Views: `web/src/views/item-details-view.ts` (интегрировать фото)
- Components: `web/src/components/` (добавить photo-uploader, photo-gallery)
- CSS: `web/src/styles/index.css` (добавить стили фото)
- Паттерн: `element()` для DOM, `on()` для событий, abort controller на destroy

### Требования:
- TypeScript strict, без non-null assertions (!)
- ESLint проходит (`npm run lint`)
- Build проходит (`npm run build`)
- TDD — тесты для логики очереди загрузки
- Follow existing code style (BEM CSS, class-based views, cleanup array)

**Начни с чтения существующих файлов, затем реализуй Этап 4. После завершения — commit & push.**