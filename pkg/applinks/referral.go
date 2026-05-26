package applinks

import "strings"

// NormalizeBotUsername strips leading @ and surrounding spaces.
func NormalizeBotUsername(username string) string {
	return strings.TrimPrefix(strings.TrimSpace(username), "@")
}

// ReferralLinkBase returns https://t.me/{bot}?start= for building referral links.
func ReferralLinkBase(botUsername string) string {
	u := NormalizeBotUsername(botUsername)
	if u == "" {
		return ""
	}
	return "https://t.me/" + u + "?start="
}

// ReferralLink builds a full referral deep link for the given referrer telegram_id.
func ReferralLink(botUsername, referrerTelegramID string) string {
	if referrerTelegramID == "" {
		return ""
	}
	base := ReferralLinkBase(botUsername)
	if base == "" {
		return ""
	}
	return base + referrerTelegramID
}
