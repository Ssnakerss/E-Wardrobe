# Web Client Implementation Plan — Wardrobe

**Дата:** 2026-11-24  
**Версия плана:** 1.0  
**Цель:** разработать полноценный веб-клиент Wardrobe на HTML, CSS и TypeScript без UI-фреймворка  
**Сборка:** Vite  
**Production:** статическая сборка встраивается в Go backend через `go:embed` и раздаётся с того же origin, что и API

---

## 1. Назначение и границы клиента

Веб-клиент должен позволять пользователю полностью управлять гардеробом:

- регистрироваться по приглашению;
- входить и выходить из аккаунта;
- восстанавливать доступ по одноразовой recovery-ссылке;
- просматривать и редактировать дерево категорий;
- создавать, просматривать, изменять, архивировать и удалять вещи;
- искать вещи и применять фильтры;
- переключаться между карточками и таблицей;
- загружать, просматривать, удалять и упорядочивать фотографии;
- использовать автоподсказки мест хранения;
- обрабатывать конфликты версий без молчаливой перезаписи данных.

Веб-клиент является **онлайн-клиентом**.

Не реализуются:

- offline-first;
- IndexedDB как основная база;
- outbox;
- фоновая синхронизация;
- Service Worker и офлайн-кэш приложения;
- мобильный push/pull sync engine.

Изменения, сделанные через веб, автоматически попадают в серверный sync journal. Отдельная синхронизация на стороне веб-клиента не требуется.

---

## 2. Фиксированные технические решения

| Область | Решение |
|---|---|
| Язык | TypeScript со `strict: true` |
| Разметка | Семантический HTML и элементы `<template>` |
| Стили | Нативный CSS, custom properties, CSS Layers |
| UI-фреймворк | Не используется |
| Сборка | Vite |
| Маршрутизация | Небольшой собственный History API router |
| HTTP | Нативный `fetch`, единый типизированный API client |
| API types | Генерация из `docs/openapi.yaml` через `openapi-typescript` |
| Состояние | Локальное состояние views и небольшой typed store для auth/UI |
| Формы | Нативные HTML forms, Constraint Validation API + доменная валидация |
| Unit/DOM tests | Vitest + jsdom |
| E2E | Playwright |
| Accessibility tests | `@axe-core/playwright` и ручные keyboard-сценарии |
| Пакетный менеджер | npm с зафиксированным `package-lock.json` |
| Production deployment | Статические файлы внутри Go backend, без Node.js |

Активная LTS-версия Node.js фиксируется через `.node-version`, `.nvmrc` и `engines` в `package.json`.

---

## 3. Предварительные требования к backend и OpenAPI

До начала активной frontend-разработки необходимо привести API-контракт в согласованное состояние.

### 3.1. Добавить недостающие контракты

- `POST /api/v1/auth/recovery/validate`.
- `POST /api/v1/auth/recovery/complete`.
- `PUT /api/v1/items/{itemId}/photos/order`.
- Точный контракт архивирования категорий, если оно входит в MVP.
- Точный состав ошибок `409 version_conflict`.
- Состав `current snapshot`, если он возвращается вместе с конфликтом.
- Коды ошибок CSRF, validation, invitation, recovery и upload.
- Ограничения длины текстовых полей.
- Ограничения `purchase_year`.
- Правила очистки nullable-полей через JSON `null`.

### 3.2. Проверить nullable-схемы

OpenAPI должен однозначно различать:

- отсутствие поля;
- `null`;
- конкретное значение.

Особое внимание:

- `season`;
- `brand`;
- `description`;
- `purchase_price`;
- `price_currency`;
- `last_worn_on`;
- `storage_location`;
- `parent_id`;
- `archived_at`.

### 3.3. Подтвердить поведение аутентификации

Для web client:

- backend устанавливает HttpOnly session cookie;
- JavaScript не получает и не сохраняет session token;
- `register` и `login` возвращают CSRF token либо клиент получает его через `/auth/csrf`;
- CSRF token передаётся в отдельном заголовке, например `X-CSRF-Token`;
- logout требует CSRF;
- cookie-authenticated unsafe requests проверяют CSRF;
- production работает same-origin, открытый CORS не нужен.

### 3.4. Источник истины

`docs/openapi.yaml` является источником API-типов.

Запрещено вручную поддерживать второй независимый набор интерфейсов `Item`, `Category`, `Photo` и `User`.

Если frontend обнаруживает расхождение, сначала исправляется контракт, затем повторно генерируются типы.

---

## 4. Структура проекта

```text
web/
├── public/
│   ├── favicon.svg
│   └── robots.txt
│
├── src/
│   ├── main.ts
│   │
│   ├── app/
│   │   ├── app.ts
│   │   ├── router.ts
│   │   ├── routes.ts
│   │   ├── auth-store.ts
│   │   ├── ui-store.ts
│   │   ├── config.ts
│   │   └── lifecycle.ts
│   │
│   ├── api/
│   │   ├── client.ts
│   │   ├── errors.ts
│   │   ├── csrf.ts
│   │   ├── auth.ts
│   │   ├── categories.ts
│   │   ├── items.ts
│   │   ├── photos.ts
│   │   ├── storage-locations.ts
│   │   └── generated/
│   │       └── schema.d.ts
│   │
│   ├── domain/
│   │   ├── category-tree.ts
│   │   ├── item.ts
│   │   ├── money.ts
│   │   ├── optional.ts
│   │   ├── validation.ts
│   │   └── formatters.ts
│   │
│   ├── views/
│   │   ├── login-view.ts
│   │   ├── register-view.ts
│   │   ├── recovery-view.ts
│   │   ├── items-view.ts
│   │   ├── item-details-view.ts
│   │   ├── item-form-view.ts
│   │   ├── categories-view.ts
│   │   ├── archive-view.ts
│   │   └── not-found-view.ts
│   │
│   ├── components/
│   │   ├── app-header.ts
│   │   ├── category-nav.ts
│   │   ├── category-select.ts
│   │   ├── item-card.ts
│   │   ├── item-table.ts
│   │   ├── item-filters.ts
│   │   ├── item-form.ts
│   │   ├── photo-gallery.ts
│   │   ├── photo-uploader.ts
│   │   ├── pagination.ts
│   │   ├── confirm-dialog.ts
│   │   ├── conflict-dialog.ts
│   │   ├── loading-state.ts
│   │   ├── empty-state.ts
│   │   ├── error-state.ts
│   │   └── toast.ts
│   │
│   ├── dom/
│   │   ├── element.ts
│   │   ├── template.ts
│   │   ├── events.ts
│   │   └── focus.ts
│   │
│   ├── styles/
│   │   ├── index.css
│   │   ├── tokens.css
│   │   ├── reset.css
│   │   ├── layout.css
│   │   ├── components.css
│   │   ├── forms.css
│   │   └── utilities.css
│   │
│   └── tests/
│       ├── setup.ts
│       ├── fixtures.ts
│       └── api-mocks.ts
│
├── e2e/
│   ├── auth.spec.ts
│   ├── categories.spec.ts
│   ├── items.spec.ts
│   ├── photos.spec.ts
│   ├── conflicts.spec.ts
│   └── accessibility.spec.ts
│
├── scripts/
│   └── check-openapi-diff.mjs
│
├── index.html
├── package.json
├── package-lock.json
├── tsconfig.json
├── vite.config.ts
├── vitest.config.ts
├── playwright.config.ts
├── eslint.config.js
├── .node-version
├── .nvmrc
└── .gitignore
```

Сгенерированный `src/api/generated/schema.d.ts` не редактируется вручную.

---

## 5. Архитектура приложения

### 5.1. Общий поток

```text
Browser URL
    │
    ▼
History API Router
    │
    ▼
Auth Guard / Route Loader
    │
    ▼
View
    │
    ├── API Client ───────► Go Backend
    │
    ├── Domain validation
    │
    └── DOM Components
```

Views управляют загрузкой данных и состоянием экрана. Components отвечают за ограниченную часть UI и не выполняют произвольные сетевые запросы без явной передачи зависимостей.

### 5.2. Жизненный цикл view

Каждый view реализует единый жизненный цикл:

```ts
interface View {
    mount(container: HTMLElement): Promise<void> | void;
    destroy(): void;
}
```

При `destroy()`:

- отменяются активные `fetch` через `AbortController`;
- удаляются подписки;
- удаляются глобальные обработчики;
- освобождаются object URL;
- останавливаются timers;
- сбрасываются незавершённые upload-операции согласно правилам uploader.

Это предотвращает обновление уже удалённого DOM и утечки обработчиков.

### 5.3. DOM без UI-фреймворка

Используются:

- `<template>` для статической разметки;
- `document.createElement`;
- небольшой typed helper `element()`;
- `textContent` для всех пользовательских и серверных строк;
- event delegation для списков;
- нативные `<button>`, `<a>`, `<form>`, `<input>`, `<select>` и `<dialog>`.

Запрещено вставлять API-данные через небезопасный `innerHTML`.

Статический `innerHTML`, если он вообще используется, должен находиться только в контролируемых шаблонах без интерполяции внешних данных.

### 5.4. Состояние

Не создаётся универсальный state framework.

Разделяются:

- auth state: пользователь, статус bootstrap, CSRF state;
- route state: текущий URL и параметры;
- view state: загрузка, ошибка, данные, черновик формы;
- UI state: карточки/таблица, открытый dialog, toast;
- server state: получено из API и не считается локальной offline-копией.

Предпочтение отображения списка можно хранить в `localStorage`, поскольку это не секрет и не доменные данные. Auth tokens, приглашения и содержимое гардероба там не хранятся.

---

## 6. Маршруты

### Публичные

```text
/login
/register?invite=<token>
/recovery?token=<token>
```

### Аутентифицированные

```text
/items
/items/new
/items/:itemId
/items/:itemId/edit
/categories
/archive
```

### Служебные

```text
/404
```

Неизвестный маршрут показывает веб-страницу 404. Неизвестный `/api/*` никогда не должен получать `index.html`.

### Поведение router

- Перехватываются только same-origin ссылки с явной SPA-навигацией.
- Не перехватываются Ctrl/Cmd/Shift/Alt-click, middle click, download и external links.
- Поддерживаются `pushState`, `replaceState` и `popstate`.
- После навигации обновляются `document.title` и focus на основном заголовке.
- Сохраняется корректная браузерная история.
- При 401 выполняется редирект на `/login` с безопасным локальным return URL.
- Внешний return URL не принимается.
- Несохранённая форма запрашивает подтверждение перед переходом.
- `beforeunload` используется только при наличии реальных несохранённых изменений.

---

## 7. API client

### 7.1. Единая точка HTTP-взаимодействия

Все запросы проходят через общий client, который отвечает за:

- `credentials: "same-origin"`;
- JSON headers;
- CSRF для unsafe methods;
- `AbortSignal`;
- запрет ручной установки `Content-Type` для `FormData`;
- обработку `204 No Content`;
- разбор typed success response;
- разбор API error response;
- преобразование сетевых ошибок;
- обработку 401;
- однократное обновление CSRF при соответствующей ошибке;
- отсутствие бесконечных retry.

### 7.2. Тип ошибки

```ts
type ApiErrorKind =
    | "network"
    | "unauthorized"
    | "forbidden"
    | "not_found"
    | "version_conflict"
    | "gone"
    | "payload_too_large"
    | "validation"
    | "rate_limited"
    | "server"
    | "invalid_response";
```

В ошибке сохраняются:

- HTTP status;
- серверный `code`;
- безопасное сообщение;
- field details;
- `Retry-After`;
- current entity/version, если они предусмотрены контрактом.

### 7.3. Правила обработки

- Сетевая ошибка не означает logout.
- 401 очищает auth state и запускает единый auth redirect, без redirect loop.
- 404 чужого ресурса отображается как обычное «не найдено».
- 409 не перезаписывает форму автоматически.
- 410 используется для просроченного или использованного приглашения/recovery.
- 413 показывает понятное сообщение о размере файла.
- 422 сопоставляется с конкретными полями, если backend вернул details.
- 429 учитывает `Retry-After`.
- 5xx показывает retry без потери введённых данных.

### 7.4. Runtime boundary

TypeScript-типы не проверяют ответ во время выполнения.

Минимально проверяются:

- ожидаемый `Content-Type`;
- JSON object вместо неожиданного scalar;
- обязательные поля критичных auth-ответов;
- UUID и version при операциях, где их отсутствие ломает дальнейшую работу.

При неожиданном ответе возвращается `invalid_response`, а не происходит молчаливая работа с повреждёнными данными.

---

## 8. Аутентификация, приглашения и recovery

### 8.1. Bootstrap приложения

При загрузке:

1. Отображается минимальный app shell.
2. Для защищённого маршрута вызывается `GET /auth/me`.
3. При успехе сохраняется пользователь и загружается CSRF token при необходимости.
4. При 401 открывается публичный маршрут.
5. При сетевой ошибке показывается ошибка доступности сервера, а не форма login как при отсутствии сессии.

### 8.2. Регистрация

- Invitation token извлекается из URL.
- Выполняется validate request.
- Пользователь видит состояние проверки, привязанный email при наличии и срок действия.
- Email не изменяется, если приглашение жёстко привязано к нему.
- Пароль имеет понятные требования и подтверждение.
- Повторный submit блокируется.
- После успеха выполняется переход к гардеробу.
- Использованный invitation token удаляется из URL и временного хранилища.

Для защиты от утечки:

- production использует `Referrer-Policy: no-referrer`;
- invitation/recovery token не отправляется в analytics;
- token не пишется в console/log;
- после чтения токен может временно храниться в `sessionStorage`, а query удаляется через `history.replaceState`;
- `localStorage` не используется.

### 8.3. Login и logout

Login:

- нормализует email по правилам backend;
- сохраняет безопасный return URL;
- не показывает, существует ли конкретный email, сверх того, что разрешает API;
- обрабатывает 429.

Logout:

- выполняется POST с CSRF;
- локальное auth state очищается только по определённому результату либо при явном fallback-сценарии;
- после выхода защищённые данные удаляются из DOM;
- выполняется переход на `/login`.

### 8.4. Recovery

- Validate и complete реализуются после добавления endpoints в OpenAPI.
- Recovery token обрабатывается так же осторожно, как invitation token.
- После успешной смены пароля старые frontend auth state и CSRF state сбрасываются.
- Пользователь перенаправляется на login с понятным сообщением.

---

## 9. Основные экраны

### 9.1. Список вещей

Содержит:

- заголовок и кнопку «Добавить вещь»;
- поиск;
- дерево или выбор категории;
- фильтры;
- сортировку;
- переключатель карточки/таблица;
- список результатов;
- cursor pagination;
- skeleton loading;
- empty state;
- error state с retry.

Карточка показывает:

- thumbnail;
- название;
- бренд;
- категорию;
- сезон;
- состояние;
- место хранения;
- признак архива, если применимо.

Таблица использует настоящие `<table>`, `<th scope="col">` и доступные controls.

### 9.2. Фильтры и URL

Состояние списка отражается в `URLSearchParams`:

```text
q
category_id
include_descendants
status
season
condition_rating
preference_rating
archived
sort
cursor
limit
```

Правила:

- изменение фильтра сбрасывает cursor;
- поиск использует debounce около 300 мс;
- предыдущий запрос отменяется;
- дополнительно применяется защита от опоздавшего ответа;
- кнопка «Сбросить фильтры» возвращает безопасные значения по умолчанию;
- Back/Forward восстанавливает состояние;
- невалидные query values нормализуются или игнорируются с понятным поведением.

### 9.3. Карточка вещи

- Галерея фотографий.
- Все атрибуты вещи.
- Цена с валютой.
- Категория.
- Последняя дата носки.
- Место хранения.
- Кнопки редактирования, архивирования и удаления.
- Быстрые действия:
  - «Носил сегодня»;
  - изменить место хранения.

Быстрые действия используют обычный versioned update. При конфликте открывается стандартный conflict flow.

### 9.4. Создание и редактирование

Поля:

- название;
- категория;
- бренд;
- описание;
- год покупки;
- цена;
- валюта;
- место покупки;
- состояние;
- личная оценка;
- дата последней носки;
- место хранения;
- сезон;
- цвет;
- размер;
- статус.

Поведение:

- при создании обязательные значения задаются явно;
- при редактировании отправляется только изменённый набор полей и `base_version`;
- очистка nullable-поля передаётся как `null`;
- неизменённые поля не добавляются в patch;
- серверные field errors отображаются рядом с полями;
- введённые значения не теряются при сетевой ошибке;
- повторный submit блокируется;
- успешный ответ обновляет локальную серверную версию сущности.

### 9.5. Категории

- Вложенный список категорий.
- Создание корневой и дочерней категории.
- Переименование.
- Перемещение.
- Сортировка через `sort_order`.
- Архивирование, если поддерживается контрактом.
- Удаление с подтверждением.

Для MVP предпочтителен семантический вложенный список с кнопками раскрытия, а не ARIA `tree`, требующий сложной клавиатурной модели.

UI не должен предлагать заведомо недопустимые parent variants:

- саму категорию;
- её потомков;
- удалённые категории.

Серверная проверка всё равно остаётся обязательной.

При `category_in_use` показывается объяснение и предлагается сначала переместить вещи/дочерние категории либо использовать архивирование.

---

## 10. Доменные правила frontend

### 10.1. Цена и валюта

- Цена — целые основные единицы валюты.
- По умолчанию выбран `RUB`.
- Начальный список: `RUB`, `USD`, `EUR`, `KZT`, `CNY`.
- Значение хранится и передаётся строкой.
- Не использовать JavaScript `number` для представления цены.
- Не выполнять преобразование через `parseFloat`.
- Копейки и центы не поддерживаются.
- Отрицательные и дробные значения отклоняются.
- Цена и валюта заполняются вместе.
- При очистке цены очищается и валюта в patch согласно контракту.
- При форматировании используется `Intl.NumberFormat`, но исходная строка не теряет точность.

### 10.2. Место хранения

Используется свободный текст с подсказками.

MVP-реализация:

- native `<input>` с `<datalist>`;
- debounce запросов;
- `GET /api/v1/storage-locations?q=...&limit=10`;
- отмена предыдущего запроса;
- подсказки не блокируют ввод;
- ошибка подсказок не мешает сохранить вещь;
- пользователь может ввести произвольное новое значение.

### 10.3. Даты

- `last_worn_on` использует `<input type="date">`.
- В API передаётся `YYYY-MM-DD`.
- Никаких произвольных UTC-преобразований через `new Date("YYYY-MM-DD")` для сохранения даты.
- Timestamps форматируются отдельно от calendar dates.

### 10.4. Archive и delete

- Archive не является удалением.
- Удаление требует подтверждения и `base_version`.
- После удаления выполняется навигация к списку и инвалидируются связанные данные.
- Ошибка конфликта не трактуется как успешное удаление.

---

## 11. Фотографии

### 11.1. Загрузка из веба

Используется:

```text
POST /api/v1/items/{itemId}/photos
```

с `multipart/form-data`.

Правила:

- поддерживаются JPEG-файлы;
- максимум 12 MiB на файл;
- учитывается серверный лимит количества фотографий;
- клиентская проверка является UX, сервер остаётся источником истины;
- `Content-Type` для multipart вручную не устанавливается;
- CSRF header добавляется;
- повторный submit одной очереди блокируется.

### 11.2. Создание вещи с фотографиями

Последовательность:

1. Сохраняется карточка вещи.
2. После получения `itemId` запускается очередь фотографий.
3. Вещь не откатывается, если одна фотография не загрузилась.
4. Для каждой фотографии отображается отдельный статус.
5. Пользователь может повторить неудачные загрузки.
6. После завершения загружается актуальная карточка с серверными photo versions.

Допустимый упрощённый MVP-flow: сначала создать вещь без фото, затем открыть менеджер фотографий. Выбранный flow фиксируется до реализации и не должен различаться между похожими экранами.

### 11.3. Preview и галерея

- Preview создаётся через `URL.createObjectURL`.
- Object URL отзывается после удаления preview или уничтожения view.
- Используются `loading="lazy"`, размеры изображения и понятный `alt`.
- Full-size просмотр реализуется через доступный `<dialog>`.
- Focus возвращается на элемент, открывший dialog.
- Escape закрывает просмотр.
- Thumbnail и original запрашиваются раздельно.

### 11.4. Удаление и порядок

- Удаление требует `base_version` фотографии.
- Порядок меняется кнопками «влево/вправо» или «вверх/вниз»; drag-and-drop может быть дополнительным UX, но не единственным способом.
- После изменения порядка отправляется полный ordered ID list и версии согласно контракту.
- Обложкой является фотография с минимальным `sort_order`.
- При конфликте порядок перечитывается, локальная операция не считается успешной.

### 11.5. Статусы

```text
queued
uploading
processing
ready
failed
```

Ошибка фотографии не удаляет карточку вещи и не блокирует независимые файлы.

---

## 12. Конфликты версий

При `409 version_conflict`:

1. Черновик пользователя сохраняется.
2. Показывается conflict dialog или отдельная панель.
3. Загружается актуальная серверная версия, если она не была возвращена в ошибке.
4. Пользователь видит changed fields и актуальные значения.
5. Доступны действия:
   - загрузить серверную версию и отказаться от черновика;
   - оставить свои значения и повторить update с новой `base_version`;
   - вернуться к ручному сравнению.

Запрещено:

- молча повторять запрос с новой версией;
- автоматически выбирать последнюю запись;
- очищать форму до явного выбора пользователя;
- показывать успех, если сервер изменение не принял.

---

## 13. CSS, адаптивность и accessibility

### 13.1. CSS

- Design tokens для цветов, spacing, typography, radius и shadows.
- CSS Layers для упорядочивания reset/base/components/utilities.
- Mobile-first layout.
- Grid для списков, Flexbox для controls.
- Container/media queries только там, где они упрощают компонент.
- Нет inline event handlers.
- Нет зависимости от CSS-in-JS.

### 13.2. Breakpoints

Точные breakpoints определяются по содержимому, а не по списку устройств. Минимально проверяются:

- узкий телефон;
- широкий телефон;
- планшет;
- стандартный desktop;
- широкий desktop.

### 13.3. Accessibility

- Один основной `<h1>` на view.
- Все inputs имеют связанные `<label>`.
- Ошибки связываются через `aria-describedby`.
- Используются `aria-live` для асинхронных статусов.
- Вся функциональность доступна с клавиатуры.
- Focus visible не отключён.
- После route change focus управляется явно.
- Dialog имеет название и корректный возврат focus.
- Контраст соответствует WCAG AA.
- Поддерживается `prefers-reduced-motion`.
- Цвет не является единственным носителем ошибки или статуса.
- Click targets на узких экранах имеют достаточный размер.

Целевые браузеры: последние две стабильные версии Chrome/Edge, Firefox и Safari. Internet Explorer не поддерживается.

---

## 14. Безопасность

Frontend-меры:

- Все API-строки выводятся через `textContent`.
- Нет `eval` и динамического `Function`.
- Нет inline `onclick`.
- Нет хранения session token в JavaScript storage.
- Invitation и recovery tokens не попадают в analytics/logs.
- Нет небезопасного open redirect через return URL.
- Object URLs своевременно освобождаются.
- Не доверять расширению и MIME, указанному браузером, как серверной проверке.
- Не отображать серверный HTML как HTML.

Production security headers настраиваются на backend/static server:

```text
Content-Security-Policy
Referrer-Policy: no-referrer
X-Content-Type-Options: nosniff
frame-ancestors 'none'
```

CSP должен разрешать только необходимые same-origin ресурсы и object/blob images, если они реально используются для preview. Dev CSP не копируется механически из production, поскольку Vite dev server имеет собственные websocket/HMR-требования.

---

# 15. Пошаговый план реализации

## Этап 0. Контракты, scaffold и качество сборки

**Оценка:** 2–3 дня.

### Задачи

- Согласовать пробелы OpenAPI из раздела 3.
- Создать Vite TypeScript project без UI-фреймворка.
- Включить strict TypeScript.
- Настроить npm lockfile и фиксированную Node.js LTS.
- Настроить ESLint и formatting.
- Настроить Vitest и Playwright scaffold.
- Настроить генерацию OpenAPI types.
- Добавить проверку, что generated files соответствуют OpenAPI.
- Настроить npm scripts:

```text
dev
build
preview
typecheck
lint
format:check
test
test:coverage
test:e2e
openapi:generate
openapi:check
check
```

- `build` должен выполнять `tsc --noEmit` до `vite build`, поскольку Vite сам не заменяет typecheck.
- Настроить Vite dev proxy `/api` на backend.
- Использовать same-origin API base по умолчанию.
- Не помещать секреты в `VITE_*`: всё с таким prefix попадает в клиентскую сборку.

### Критерии готовности

- `npm ci` работает на чистом checkout.
- `npm run check` проходит.
- Dev server открывает минимальную страницу.
- `/api` проксируется на backend.
- Production build создаётся без ошибок.
- Изменение OpenAPI без повторной генерации обнаруживается в CI.

---

## Этап 1. App shell, router, API client и auth

**Оценка:** 4–5 дней.

### Задачи

- Реализовать DOM helpers и lifecycle views.
- Реализовать router и route guards.
- Реализовать typed API client.
- Реализовать error normalization.
- Реализовать auth store.
- Реализовать CSRF lifecycle.
- Сделать app shell:
  - header;
  - основную навигацию;
  - main;
  - toast/status region;
  - footer при необходимости.
- Реализовать invitation registration.
- Реализовать login/logout.
- Реализовать recovery.
- Добавить базовые tokens и responsive layout.
- Добавить loading/error/empty components.

### Тесты

- Публичный маршрут не требует `/auth/me` для отображения.
- Защищённый маршрут корректно обрабатывает 401.
- Network error при bootstrap не отображается как «неправильный пароль».
- CSRF обновляется не более одного раза на запрос.
- Logout отправляет CSRF.
- Return URL не допускает внешний redirect.
- Invitation/recovery token удаляется из URL согласно выбранному flow.
- Browser Back/Forward не ломает auth state.
- Несколько одновременных 401 не создают redirect loop.

### Критерий готовности

Пользователь может зарегистрироваться по приглашению, войти, обновить страницу, остаться авторизованным и выйти.

---

## Этап 2. Категории и навигация по гардеробу

**Оценка:** 3–4 дня.

### Задачи

- Получить flat list категорий.
- Построить дерево чистой доменной функцией.
- Обработать отсутствующего parent без бесконечного цикла UI.
- Реализовать category navigation.
- Реализовать create/rename/move/delete.
- Добавить versioned updates.
- Добавить понятную обработку `category_in_use`.
- Реализовать reusable category select для формы вещи.
- Добавить архивирование после фиксации контракта.

### Тесты

- Корректное построение дерева и сортировка.
- Защита frontend от malformed/cyclic fixture.
- Нельзя выбрать потомка как нового parent.
- 404 чужой категории не раскрывает дополнительную информацию.
- 409 сохраняет введённые изменения.
- После mutation связанные списки обновляются без полной перезагрузки страницы.

### Критерий готовности

Пользователь полностью управляет деревом категорий и может использовать его в навигации.

---

## Этап 3. Вещи: список, детали, формы и архив

**Оценка:** 6–8 дней.

### Задачи

- Реализовать список карточек.
- Реализовать таблицу.
- Реализовать URL-driven filters/search/sort.
- Реализовать cursor pagination.
- Реализовать item details.
- Реализовать create/edit forms.
- Реализовать optional/null patch semantics.
- Реализовать цену строкой и выбор валюты.
- Реализовать storage suggestions.
- Реализовать archive/unarchive.
- Реализовать delete.
- Реализовать быстрые действия.
- Реализовать dirty-form guard.
- Реализовать общий conflict flow.

### Тесты

- Фильтры формируют правильный query.
- Смена фильтра сбрасывает cursor.
- Устаревший search response не перезаписывает новый.
- Пагинация сохраняет сортировку и фильтры.
- Пустая цена не превращается в `0`.
- Дробная цена отклоняется.
- `null` очищает поле, отсутствие поля его не меняет.
- Ошибка сети не очищает форму.
- Двойной click не создаёт две вещи.
- Archive и delete имеют разные последствия.
- 409 не приводит к silent overwrite.

### Критерий готовности

Веб-клиент позволяет полноценно управлять вещами без фотографий.

---

## Этап 4. Фотографии

**Оценка:** 4–5 дней.

### Задачи

- Реализовать выбор JPEG-файлов.
- Добавить preview и очистку object URLs.
- Реализовать multipart upload.
- Реализовать очередь с ограниченным concurrency.
- Добавить per-file status/retry.
- Реализовать gallery/lightbox.
- Реализовать delete.
- Реализовать reorder.
- Реализовать обложку.
- Обновлять photo versions после каждой операции.
- Обрабатывать 413, 422, 409 и network errors отдельно.

### Тесты

- Неподдерживаемый файл отклоняется в UI.
- Oversized файл показывает понятную ошибку.
- Один failed upload не отменяет остальные.
- Retry не создаёт непреднамеренный дубликат в UI.
- Удаление и reorder используют актуальные versions.
- После навигации object URLs отзываются.
- Форма вещи остаётся сохранённой при ошибке фотографии.
- Keyboard может изменить порядок и закрыть галерею.

### Критерий готовности

Пользователь может загрузить несколько фотографий, просмотреть их, изменить порядок и удалить.

---

## Этап 5. UX, accessibility и адаптивная стабилизация

**Оценка:** 3–4 дня.

### Задачи

- Проверить все loading/empty/error states.
- Довести responsive table/cards.
- Проверить длинные названия, описания и места хранения.
- Проверить отсутствие фото и битые thumbnail responses.
- Проверить keyboard-only сценарии.
- Добавить axe checks.
- Проверить focus после route/dialog/mutation.
- Проверить reduced motion.
- Проверить zoom до 200%.
- Устранить layout shift у изображений.
- Проверить русские и потенциально длинные английские строки.

### Критерий готовности

Основные сценарии доступны с клавиатуры, не имеют критических axe-нарушений и корректно отображаются на целевых размерах экрана.

---

## Этап 6. Интеграция с Go backend и production static hosting

**Оценка:** 2–3 дня.

### Задачи

Production build pipeline:

```text
npm ci
npm run check
npm run build
очистить backend/internal/webui/dist
скопировать web/dist в backend/internal/webui/dist
go build ./...
```

Правила:

- `backend/internal/webui/dist` является производным артефактом, его не редактируют вручную.
- Для `go:embed` должен существовать корректный placeholder до первой реальной сборки.
- Vite `base` соответствует корневой раздаче `/`.
- Hashed assets кэшируются долго.
- `index.html` не кэшируется надолго.
- SPA fallback применяется только к frontend routes.
- `/api/*`, health endpoints и служебные маршруты не попадают под fallback.
- Несуществующий asset возвращает 404, а не `index.html`.
- Node.js в production container не требуется.
- Проверяются production CSP и referrer policy.
- Проверяется HTTPS и Secure cookie.

### Критерий готовности

Один Go binary обслуживает API и production web client, включая прямые переходы и refresh на `/items/:id`.

---

## Этап 7. E2E, release hardening и документация

**Оценка:** 4–5 дней.

### E2E против реального backend и PostgreSQL

1. Создание приглашения через CLI и регистрация.
2. Повторное использование приглашения.
3. Login/logout и expired session.
4. Recovery flow.
5. Создание и перемещение категории.
6. Создание вещи со всеми полями.
7. Проверка цены `12990 RUB` без потери строкового представления.
8. Поиск, фильтр, сортировка и pagination.
9. Редактирование и очистка nullable-полей.
10. Storage suggestions.
11. Archive/unarchive.
12. Upload нескольких JPEG.
13. Reorder и delete фотографий.
14. Version conflict с сохранением черновика.
15. Изоляция двух пользователей.
16. Прямой refresh защищённого route.
17. Keyboard и axe smoke suite.

### CI

- Unit/DOM tests на каждый pull request.
- Chromium E2E на основные сценарии.
- Firefox и WebKit перед release или по регулярному schedule.
- OpenAPI diff check.
- Production build и Go embed smoke test.
- Проверка отсутствия секретов и исходных invitation URLs в артефактах тестов.

### Документация

- `web/README.md`: установка и команды.
- Настройка dev proxy.
- Production build и копирование dist.
- Правила работы с OpenAPI.
- Как добавлять view/component.
- Как запускать E2E.
- Известные ограничения онлайн-клиента.

---

## 16. Тестовая стратегия

### Unit tests

Обязательно покрываются чистые функции:

- category tree;
- price parsing/formatting;
- optional/null patch building;
- query serialization;
- cursor reset;
- route matching;
- safe return URL;
- error normalization;
- validation;
- sorting/filter view helpers.

### DOM integration tests

- Формы и ошибки полей.
- Блокировка повторного submit.
- Race conditions поиска.
- Жизненный цикл views.
- Диалоги и focus.
- Upload statuses.
- Conflict flow.
- Очистка subscriptions и object URLs.

### E2E

E2E не заменяются mocked DOM tests. Критические flows выполняются против реальных API и PostgreSQL.

### Ручные проверки

- Реальный upload больших изображений.
- Медленная сеть.
- Offline во время запроса: ошибка отображается, но offline-работа не обещается.
- Истечение сессии во время редактирования.
- Back/Forward во время несохранённой формы.
- Safari date/input/dialog behaviour.
- Screen reader smoke test основных форм.

---

## 17. Производительность

- Route-based lazy loading для крупных views.
- Нет глобальной загрузки всех вещей.
- Серверная pagination.
- Отмена устаревших запросов.
- Lazy loading thumbnails.
- Явные width/height или aspect ratio для изображений.
- Минимум production dependencies.
- Нет UI framework и тяжёлой state-библиотеки.
- Нет base64-фотографий в DOM/JSON.
- Object URLs освобождаются.
- Source maps production публикуются только при осознанном решении.

Перед release измеряются:

- размер initial JS/CSS;
- время открытия списка;
- время переключения фильтров;
- поведение на медленной сети;
- память при длительном просмотре галереи.

---

## 18. Зависимости и порядок работ

```text
OpenAPI и auth contract
        │
        ▼
Этап 0 — scaffold
        │
        ▼
Этап 1 — shell/router/API/auth
        │
        ▼
Этап 2 — categories
        │
        ▼
Этап 3 — items
        │
        ▼
Этап 4 — photos
        │
        ▼
Этап 5 — UX/accessibility
        │
        ▼
Этап 6 — Go embed/production
        │
        ▼
Этап 7 — E2E/release
```

Возможное распараллеливание:

- CSS tokens и app shell можно разрабатывать параллельно с уточнением доменных endpoints.
- Unit tests чистых доменных функций можно писать до готового backend.
- Фото UI можно строить на contract mocks, но принимать этап только против реального API.
- Нельзя окончательно закрывать auth и conflict flow, пока соответствующие backend контракты не зафиксированы.

---

## 19. Оценка трудоёмкости

| Этап | Оценка |
|---|---:|
| 0. Контракты и scaffold | 2–3 дня |
| 1. Shell, router, API и auth | 4–5 дней |
| 2. Категории | 3–4 дня |
| 3. Вещи | 6–8 дней |
| 4. Фотографии | 4–5 дней |
| 5. UX и accessibility | 3–4 дня |
| 6. Production integration | 2–3 дня |
| 7. E2E и release hardening | 4–5 дней |
| **Итого** | **28–37 человеко-дней** |

Оценка относится к одному опытному разработчику и включает тесты. Она не включает разработку backend и мобильного приложения.

При параллельном изменении API или нестабильном backend следует предусмотреть дополнительный резерв на интеграцию.

---

## 20. Definition of Done

Web MVP считается готовым, когда:

1. Регистрация по приглашению, login, logout и recovery работают end-to-end.
2. Session token недоступен JavaScript и не хранится в web storage.
3. CSRF корректно применяется ко всем cookie-authenticated unsafe requests.
4. Категории можно создавать, изменять, перемещать и удалять согласно backend-правилам.
5. Вещи можно создавать, просматривать, редактировать, архивировать и удалять.
6. Цена остаётся целой строкой в основных единицах выбранной валюты.
7. Nullable patch semantics покрыты тестами.
8. Search, filters, sorting и cursor pagination работают без race conditions.
9. Storage suggestions работают и не блокируют свободный ввод.
10. Фотографии загружаются, просматриваются, упорядочиваются и удаляются.
11. Ошибка фотографии не приводит к потере карточки вещи.
12. Конфликт версии никогда не разрешается молчаливой перезаписью.
13. Основные экраны адаптивны и доступны с клавиатуры.
14. Нет критических accessibility-нарушений в согласованном test suite.
15. Unit, DOM и критические E2E tests проходят.
16. Production build встроен в Go binary.
17. Refresh и прямые ссылки на SPA routes работают.
18. API paths и missing assets не получают `index.html`.
19. Security headers и поведение cookies проверены в production-окружении.
20. README содержит воспроизводимые dev, test и build команды.