<<<<<<< HEAD
# Кнопка Support

Ссылка: [https://t.me/+jXI2qDR9Y-xkZTI6](https://t.me/+jXI2qDR9Y-xkZTI6) (канал **CyberMate | Community**).

**Не** открывать `https://t.me/${BOT_USERNAME}` — это чат с ботом, не support.

Реализация во фронте: `tgapp_front/src/lib/openSupport.js`, кнопка в `App.jsx` → `handleSupportClick`.
=======
# Кнопка Support → Telegram

## Ссылка

Чат/канал сообщества CyberMate: [https://t.me/+jXI2qDR9Y-xkZTI6](https://t.me/+jXI2qDR9Y-xkZTI6)

На бэкенде задаётся в `.env`:

```env
TELEGRAM_SUPPORT_INVITE_URL=https://t.me/+jXI2qDR9Y-xkZTI6
```

## API (опционально)
>>>>>>> 3489ac71c17ae6e070eec77e5b2b0b383107f257

```http
GET /v1/app/links
```

<<<<<<< HEAD
```json
{ "support_chat_url": "https://t.me/+jXI2qDR9Y-xkZTI6", "bot_username": "CyberMate_bot" }
```

```ts
export function openSupport() {
  const tg = window.Telegram?.WebApp;
  const url = "https://t.me/+jXI2qDR9Y-xkZTI6";
  if (tg?.openTelegramLink) tg.openTelegramLink(url);
  else window.open(url, "_blank");
}
```
=======
Ответ:

```json
{
  "support_chat_url": "https://t.me/+jXI2qDR9Y-xkZTI6",
  "bot_username": "CyberMate_bot"
}
```

## Реализация на фронте (Mini App)

При клике на кнопку **Support**:

```ts
const SUPPORT_URL =
  import.meta.env.VITE_SUPPORT_URL ?? "https://t.me/+jXI2qDR9Y-xkZTI6";

export function openSupport() {
  const tg = window.Telegram?.WebApp;
  if (tg?.openTelegramLink) {
    tg.openTelegramLink(SUPPORT_URL);
    return;
  }
  window.open(SUPPORT_URL, "_blank", "noopener,noreferrer");
}
```

Или загрузить URL с API при старте:

```ts
const { support_chat_url } = await fetch(`${API}/v1/app/links`).then((r) => r.json());
tg.openTelegramLink(support_chat_url);
```

**Важно:** в Telegram Mini App используйте `openTelegramLink`, а не обычный `<a target="_blank">` — иначе ссылка может не открыться внутри клиента Telegram.

## React-пример

```tsx
<button type="button" onClick={openSupport}>
  Support
</button>
```
>>>>>>> 3489ac71c17ae6e070eec77e5b2b0b383107f257
