package applinks

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/twelvepills-936/tgapp-/pkg/config"
)

<<<<<<< HEAD
// linksResponse is returned by GET /v1/app/links for the frontend (Support button, referral links).
=======
// linksResponse is returned by GET /v1/app/links for the frontend (Support button, etc.).
>>>>>>> c6b228cc8ba7fd9eaf6535effaa3c829367efd64
type linksResponse struct {
	SupportChatURL   string `json:"support_chat_url"`
	BotUsername      string `json:"bot_username,omitempty"`
	ReferralLinkBase string `json:"referral_link_base,omitempty"`
}

type referralLinkResponse struct {
	ReferralLink string `json:"referral_link"`
}

const referralLinkPathPrefix = "/v1/users/telegram/"
const referralLinkPathSuffix = "/referral-link"

// Wrap adds app link endpoints with Telegram deep links from config.
func Wrap(next http.Handler, app config.ConfigApp) http.Handler {
	botUsername := NormalizeBotUsername(app.TelegramBotUsername)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}

		switch {
		case r.URL.Path == "/v1/app/links":
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			_ = json.NewEncoder(w).Encode(linksResponse{
				SupportChatURL:   app.SupportTelegramInviteURL,
				BotUsername:      botUsername,
				ReferralLinkBase: ReferralLinkBase(botUsername),
			})
			return

		case strings.HasPrefix(r.URL.Path, referralLinkPathPrefix) &&
			strings.HasSuffix(r.URL.Path, referralLinkPathSuffix):
			telegramID := strings.TrimSuffix(
				strings.TrimPrefix(r.URL.Path, referralLinkPathPrefix),
				referralLinkPathSuffix,
			)
			if telegramID == "" {
				http.Error(w, "telegram_id required", http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			_ = json.NewEncoder(w).Encode(referralLinkResponse{
				ReferralLink: ReferralLink(botUsername, telegramID),
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}
