# Кнопка Support → Telegram

## Ссылка

Чат/канал сообщества CyberMate: [https://t.me/+jXI2qDR9Y-xkZTI6](https://t.me/+jXI2qDR9Y-xkZTI6)

**Не** открывать `https://t.me/${BOT_USERNAME}` — это чат с ботом, не support.

На бэкенде задаётся в `.env`:

```env
TELEGRAM_SUPPORT_INVITE_URL=https://t.me/+jXI2qDR9Y-xkZTI6
TELEGRAM_BOT_USERNAME=CyberMate_bot
```

## API

```http
GET /v1/app/links
```

Ответ:

```json
{
  "support_chat_url": "https://t.me/+jXI2qDR9Y-xkZTI6",
  "bot_username": "CyberMate_bot",
  "referral_link_base": "https://t.me/CyberMate_bot?start="
}
```

Готовая реферальная ссылка пользователя:

```http
GET /v1/users/telegram/{telegram_id}/referral-link
```

```json
{ "referral_link": "https://t.me/CyberMate_bot?start=123456789" }
```

## Реализация на фронте (Mini App)

### Support

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

### Реферальная ссылка

Берите `referral_link` из API профиля или соберите из `/v1/app/links`:

```ts
// Вариант 1: готовая ссылка
const { referral_link } = await fetch(
  `${API}/v1/users/telegram/${telegramUserId}/referral-link`
).then((r) => r.json());

// Вариант 2: из конфига приложения
const { referral_link_base } = await fetch(`${API}/v1/app/links`).then((r) => r.json());
const referralLink = referral_link_base + telegramUserId;
```

При регистрации `start_param` в `POST /v1/register` должен содержать `telegram_id` пригласившего.

**Важно:** в Telegram Mini App используйте `openTelegramLink`, а не обычный `<a target="_blank">` — иначе ссылка может не открыться внутри клиента Telegram.
