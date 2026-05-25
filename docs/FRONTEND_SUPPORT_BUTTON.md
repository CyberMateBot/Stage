# Кнопка Support

Ссылка: [https://t.me/+jXI2qDR9Y-xkZTI6](https://t.me/+jXI2qDR9Y-xkZTI6)

```http
GET /v1/app/links
```

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
