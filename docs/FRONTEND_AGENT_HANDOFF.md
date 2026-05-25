# Handoff: фронтенд CyberMate (смена темы)

Документ для агента/разработчика, который реализует **отдельный frontend** (Vite/React/Vue и т.д.). Backend — репозиторий `tgapp-` (Go), **без встроенной раздачи UI**.

---

## 1. Архитектура

| Слой | Репозиторий / путь | Ответственность |
|------|-------------------|-----------------|
| Backend | `d:\tgapp-` (этот репо) | API, БД, CORS, миграции |
| Frontend | **отдельный проект** (ещё не в workspace) | UI, `data-theme`, Telegram WebApp SDK |

Связь: браузер/Mini App на `http://localhost:5173` (или prod URL) → HTTP к API `http://localhost:8090` с заголовком `Origin`. Доступ разрешён только если origin указан в **`CORS_ALLOWED_ORIGINS`** в `.env` бэкенда.

---

## 2. CORS (обязательно настроить до отладки UI)

Файл: **`.env`** в корне backend (см. `pkg/config/config.go` → `LoadCORSConfig`).

```env
# Один origin
CORS_ALLOWED_ORIGINS=http://localhost:5173

# Несколько (через запятую, без пробелов или с пробелами — trim делается)
CORS_ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000

# Dev: разрешить все (не для production)
CORS_ALLOWED_ORIGINS=*
```

Дополнительно (опционально):

```env
CORS_ALLOWED_METHODS=GET,POST,PUT,PATCH,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Content-Type,Authorization
```

Middleware: `pkg/cors/middleware.go`, подключение в `cmd/service/main.go` (оборачивает gRPC-Gateway + Swagger).

**На фронте** задайте базовый URL API (не `window.location.origin` Mini App-хоста):

```env
# .env фронтенда (Vite)
VITE_API_URL=http://localhost:8090
```

---

## 3. API темы (backend уже реализован)

### 3.1 Получить профиль и текущую тему

```http
GET /v1/users/telegram/{telegram_id}
```

Ответ (фрагмент):

```json
{
  "data": {
    "id": 1,
    "name": "Ivan",
    "surname": "",
    "theme": "light"
  }
}
```

Поле `theme`: **`light`** | **`dark`**.

### 3.2 Сохранить тему

```http
PATCH /v1/users/telegram/{telegram_id}/theme
Content-Type: application/json

{ "theme": "dark" }
```

`telegram_id` в path берётся из URL; в body достаточно `{ "theme": "dark" }` (path перезапишет `telegram_id` в proto).

Ответ:

```json
{ "theme": "dark" }
```

Ошибки:

| HTTP | Код / смысл |
|------|-------------|
| 404 | Профиль не найден (`PROFILE_NOT_FOUND`) |
| 400 | Невалидная тема (`INVALID_ARGUMENT`, только `light`/`dark`) |

### 3.3 База данных

Миграция: `internal/migrations/V20250525000000__profile_ui_theme.sql`  
Колонка: `profiles.ui_theme` (`light` | `dark`, default `light`).

### 3.4 Цепочка на backend (для отладки)

```
HTTP PATCH → pkg/api (grpc-gateway) → internal/service/service.go UpdateProfileTheme
→ internal/usecase/profile.go UpdateProfileTheme
→ internal/repository/profiles.go UpdateProfileTheme
```

Константы: `internal/usecase/models/profile.go` — `ThemeLight`, `ThemeDark`.

---

## 4. Логика темы на фронте (что нужно портировать)

Эталон (vanilla): **`docs/frontend-theme-reference/`**

| Файл | Назначение |
|------|------------|
| `index.html` | Разметка переключателя «Светлая / Тёмная» |
| `app.css` | CSS-переменные для `[data-theme="light"]` и `[data-theme="dark"]` |
| `app.js` | Полный flow: localStorage, Telegram, API |

### 4.1 Константы

- `STORAGE_KEY = "cybermate-ui-theme"`
- `THEMES = ["light", "dark"]`

### 4.2 Применение темы (`applyTheme`)

1. `document.documentElement.setAttribute("data-theme", "light"|"dark")`
2. `localStorage.setItem(STORAGE_KEY, theme)`
3. Подсветка активной кнопки (класс `is-active`)
4. Telegram WebApp (если есть):
   - dark: `setHeaderColor("#0b1220")`, `setBackgroundColor("#0b1220")`
   - light: `#f4f6fb`

### 4.3 Порядок при старте (`resolveInitialTheme`)

1. Если есть валидное значение в `localStorage` → использовать его.
2. Иначе `GET /v1/users/telegram/{telegram_id}` → `data.theme`.
3. Иначе `Telegram.WebApp.colorScheme` → `dark` или `light`.
4. Fallback: `light`.

`telegram_id` = `String(Telegram.WebApp.initDataUnsafe.user.id)`.

### 4.4 При смене пользователем (`selectTheme`)

1. `applyTheme(theme)` сразу (UX).
2. `PATCH .../theme` с `{ theme }`.
3. При ошибке сети — оставить локальную тему, показать предупреждение.

### 4.5 Событие Telegram

```js
Telegram.WebApp.onEvent("themeChanged", () => {
  if (!localStorage.getItem(STORAGE_KEY)) {
    applyTheme(telegramTheme(), { silent: true });
  }
});
```

Если пользователь уже выбрал тему вручную (есть ключ в localStorage), системную смену темы Telegram **не перезаписывать**.

---

## 5. Рекомендуемая структура фронт-проекта

```
frontend/                    # новый репозиторий или папка рядом с tgapp-
├── .env
│   └── VITE_API_URL=http://localhost:8090
├── src/
│   ├── api/
│   │   └── theme.ts         # getProfileTheme, updateProfileTheme
│   ├── hooks/
│   │   └── useTheme.ts      # resolveInitialTheme, selectTheme
│   ├── styles/
│   │   └── theme.css        # из docs/frontend-theme-reference/app.css
│   └── components/
│       └── ThemeToggle.tsx
└── index.html               # script telegram-web-app.js
```

Пример `theme.ts`:

```ts
const API = import.meta.env.VITE_API_URL;

export async function getProfileTheme(telegramId: string): Promise<"light" | "dark" | null> {
  const res = await fetch(`${API}/v1/users/telegram/${encodeURIComponent(telegramId)}`);
  if (!res.ok) return null;
  const json = await res.json();
  const t = json?.data?.theme;
  return t === "light" || t === "dark" ? t : null;
}

export async function updateProfileTheme(telegramId: string, theme: "light" | "dark") {
  const res = await fetch(`${API}/v1/users/telegram/${encodeURIComponent(telegramId)}/theme`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ theme }),
  });
  if (!res.ok) throw new Error(`theme update failed: ${res.status}`);
  return res.json();
}
```

---

## 6. Чеклист для фронт-агента

- [ ] Создать отдельный frontend-проект (Vite + React/Vue — на выбор команды).
- [ ] `.env`: `VITE_API_URL` → URL backend (`8090` по умолчанию).
- [ ] В backend `.env`: добавить origin фронта в `CORS_ALLOWED_ORIGINS` (не забыть prod URL).
- [ ] Подключить `https://telegram.org/js/telegram-web-app.js` в Mini App.
- [ ] Реализовать `data-theme` + CSS variables (см. reference).
- [ ] Реализовать `useTheme` / store по flow из §4.
- [ ] Регистрация пользователя: перед темой профиль должен существовать (`POST /v1/register`), иначе PATCH theme → 404.
- [ ] Проверить в Telegram: смена темы сохраняется после перезапуска Mini App.

---

## 7. Что **удалено** из backend (не восстанавливать без согласования)

- Пакет `pkg/web/` (embed статики `/app/`) — **удалён**; UI только во фронт-репо.
- Раздача Mini App с того же порта API больше не используется.

---

## 8. Полезные команды backend

```bash
# Миграции применяются при старте Postgres / flyway в CI
go run cmd/service/main.go

# Swagger
# http://localhost:8090/swagger/
```

Proto (если меняете API): `make proto.gen` (Docker) или см. `scripts/proto-gen.sh`.

---

## 9. Контактные файлы backend для темы

| Файл |
|------|
| `api/service.proto` — RPC `UpdateProfileTheme`, поле `User.theme` |
| `internal/service/service.go` — HTTP/gRPC handlers |
| `internal/usecase/profile.go` — бизнес-логика |
| `internal/repository/profiles.go` — SQL |
| `pkg/cors/middleware.go` — CORS |
| `docs/frontend-theme-reference/*` — эталон UI |

**Задача фронт-агента:** перенести reference в целевой стек, подключить `VITE_API_URL` + CORS, встроить переключатель в экран настроек приложения.
