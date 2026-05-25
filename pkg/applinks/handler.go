package applinks

import (
	"encoding/json"
	"net/http"

	"github.com/twelvepills-936/tgapp-/pkg/config"
)

<<<<<<< HEAD
=======
// linksResponse is returned by GET /v1/app/links for the frontend (Support button, etc.).
>>>>>>> 3489ac71c17ae6e070eec77e5b2b0b383107f257
type linksResponse struct {
	SupportChatURL string `json:"support_chat_url"`
	BotUsername    string `json:"bot_username,omitempty"`
}

// Wrap adds GET /v1/app/links with Telegram deep links from config.
func Wrap(next http.Handler, app config.ConfigApp) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/v1/app/links" {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			_ = json.NewEncoder(w).Encode(linksResponse{
				SupportChatURL: app.SupportTelegramInviteURL,
				BotUsername:    app.TelegramBotUsername,
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}
