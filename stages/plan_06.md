**Продолжение реализации веб-клиента E-Wardrobe (Этап 6 из 8) — Production Hardening**

## Контекст
- **Проект:** E-Wardrobe — управление гардеробом
- **Бэкенд:** Go backend (`backend/`)
- **Фронтенд:** `web/` — чистый TypeScript + HTML + CSS, Vite, без UI-фреймворков
- **План:** `docs/web-implementation-plan.md`

## Текущий прогресс (6 этапов завершены)

### Завершённые этапы:
- **Этап 0** ✅ — Scaffold: Vite + TypeScript strict + ESLint + Prettier + Vitest + Playwright + OpenAPI types
- **Этап 1** ✅ — App shell, router, API client, auth
- **Этап 2** ✅ — Категории: дерево, CRUD, навигация
- **Этап 3** ✅ — Вещи: список, фильтры, поиск, детали, формы, архив, conflict dialog
- **Этап 4** ✅ — Фотографии: upload, gallery, lightbox, reorder
- **Этап 5** ✅ — UX: loading/empty/error states, responsive, long text handling, CSP, Service Worker, HTTPS dev

## Команды для проверки:
```
cd web && npm run check   # typecheck + lint + format + test
npm run build              # production build → dist/
```

---

# Этап 6: Production Hardening — go:embed, SPA Fallback, Caching

**Из плана (раздел 14 и 15, Этап 6):**

## Задачи:

### 1. go:embed — встроить веб-статику в Go binary
- **План:** "go:embed"
- В `backend/` добавить embed директиву:
  ```go
  //go:embed ../web/dist
  var webFS embed.FS
  ```
- Создать HTTP handler для раздачи static files:
  ```go
  func webHandler() http.Handler {
      fs := http.FS(webFS)
      return http.StripPrefix("/web/", http.FileServer(fs))
  }
  ```
- Настроить SPA fallback: все не-API запросы → `index.html`
- Убедиться, что `go build ./...` проходит

### 2. SPA Fallback — корректная маршрутизация
- **План:** "SPA fallback"
- Все запросы, которые не начинаются с `/api/`, должны возвращать `index.html`
- Это позволяет клиентскому router обрабатывать `/items`, `/categories`, `/login` и т.д.
- Реализация в `backend/cmd/server/main.go` или отдельном handler:
  ```go
  func spaFallback(next http.Handler) http.Handler {
      return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
          if !strings.HasPrefix(r.URL.Path, "/api/") {
              http.ServeFile(w, r, path.Join("dist", "index.html"))
              return
          }
          next.ServeHTTP(w, r)
      })
  }
  ```
- Проверить, что `/api/*` всё ещё маршрутизируется к Go handlers
- Проверить, что статические файлы (`/assets/*.js`, `/assets/*.css`) отдаются корректно

### 3. Caching — кэширование статики
- **План:** "Кэширование статики"
- Настроить правильные Cache-Control заголовки:
  - `index.html` — `no-cache` (чтобы получать новые версии)
  - `/assets/*.js`, `/assets/*.css` — `public, max-age=31536000, immutable` (хеш в имени файла)
  - `/favicon.svg`, `/robots.txt` — `public, max-age=86400`
- В Vite убедиться, что `build.rollupOptions.output.assetFileNames` и `chunkFileNames` используют хеширование
- Проверить, что Vite уже добавляет хеши в имена файлов (должно быть по умолчанию)

### 4. Проверить и донастроить CSP
- **План:** "CSP"
- Убедиться, что CSP мета-тег в `index.html` корректен:
  ```html
  <meta http-equiv="Content-Security-Policy" 
    content="default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; 
             img-src 'self' data: blob:; connect-src 'self'; object-src 'none'; base-uri 'self'">
  ```
- Проверить, что:
  - Нет inline-скриптов (кроме CSP-разрешённых)
  - Нет inline-стилей (кроме CSP-разрешённых)
  - `blob:` нужен для `URL.createObjectURL` (фото preview)
  - `data:` нужен для placeholder изображений
- Добавить CSP headers на бэкенде (опционально, вместо мета-тега)

### 5. Проверить HTTPS в production
- **План:** "HTTPS"
- Убедиться, что бэкенд поддерживает HTTPS:
  ```go
  log.Fatal(http.ListenAndServeTLS(":443", "cert.pem", "key.pem", handler))
  ```
- В `web/src/app/config.ts` убедиться, что API base — same-origin (не hardcoded URL)
- Проверить, что CSRF работает с Secure cookie:
  ```go
  httpOnly: true, Secure: true, SameSite: "Strict"
  ```

### 6. Referrer-Policy
- **План:** "Referrer-Policy: no-referrer"
- Добавить header на бэкенде:
  ```go
  w.Header().Set("Referrer-Policy", "no-referrer")
  ```
- Или мета-тег в `index.html`:
  ```html
  <meta name="referrer" content="no-referrer">
  ```

### 7. Проверить production build
- **План:** "build должен выполнять tsc --noEmit до vite build"
- Убедиться, что `npm run build` (уже настроен в package.json)
- Проверить размер бандла:
  - Общее size < 500 KB (gzip)
  - Main chunk < 100 KB (gzip)
  - Lazy chunks < 50 KB (gzip) каждый
- Если size большой — проверить, что lazy loading работает

### 8. Добавить health check endpoint
- **План:** "health check"
- Добавить `GET /health` в бэкенд:
  ```go
  func healthHandler() http.HandlerFunc {
      return func(w http.ResponseWriter, r *http.Request) {
          w.Header().Set("Content-Type", "application/json")
          w.WriteHeader(http.StatusOK)
          w.Write([]byte(`{"status":"ok"}`))
      }
  }
  ```
- Используется для health checks в Kubernetes/Docker

### 9. Добавить X-Frame-Options
- **План:** "X-Frame-Options: DENY"
- Добавить header на бэкенде:
  ```go
  w.Header().Set("X-Frame-Options", "DENY")
  ```
- Или через CSP: `frame-ancestors 'none'`

---

## Файлы для создания/изменения:

### Backend:
1. **`backend/internal/handler/web.go`** — новый файл: go:embed, webHandler, spaFallback
2. **`backend/cmd/server/main.go`** — интегрировать webHandler, SPA fallback, CSP headers, Referrer-Policy, X-Frame-Options
3. **`backend/internal/handler/health.go`** — health check endpoint

### Frontend:
4. **`web/index.html`** — проверить CSP meta tag, referrer meta tag
5. **`web/vite.config.ts`** — проверить assetFileNames с хешированием
6. **`web/src/app/config.ts`** — убедиться что API base — same-origin

---

## Ключевые требования:
- `go build ./...` проходит без ошибок
- `go run ./cmd/server` запускает сервер с встроенной статикой
- `GET /api/v1/auth/me` работает
- `GET /items` возвращает `index.html` (SPA fallback)
- `GET /assets/app.js` возвращает JS файл
- `GET /health` возвращает `{"status":"ok"}`
- CSP не блокирует загрузку ресурсов
- Cache-Control заголовки корректны
- HTTPS работает (если есть cert)
- Referrer-Policy: no-referrer
- X-Frame-Options: DENY

## Что НЕ делать:
- Не менять API contract
- Не менять логику views
- Не добавлять новые зависимости
- Не менять CSS/JS логику

---

## Порядок выполнения:
1. Проверить текущее состояние: `cd web && npm run check && npm run build`
2. Проверить структуру backend: `ls backend/cmd/server/`, `cat backend/cmd/server/main.go`
3. Создать `backend/internal/handler/web.go` с go:embed
4. Создать `backend/internal/handler/health.go` с health check
5. Интегрировать в `backend/cmd/server/main.go`
6. Проверить CSP в `web/index.html`
7. Проверить caching в `web/vite.config.ts`
8. Запустить `go build ./...` и `go run ./cmd/server`
9. Протестировать: `/api/`, `/`, `/assets/`, `/health`
10. Commit & push

---

**Начни с чтения текущего состояния backend и frontend, затем реализуй Этап 6 поэтапно. После завершения — commit & push.**