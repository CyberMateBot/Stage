package applinks

import "strings"

const DefaultReferralParamPrefix = "ref_"

// NormalizeBotUsername strips leading @ and surrounding spaces.
func NormalizeBotUsername(username string) string {
	return strings.TrimPrefix(strings.TrimSpace(username), "@")
}

// ReferralLinkBase returns https://t.me/{bot}?startapp={prefix} for Mini App referral links.
func ReferralLinkBase(botUsername, paramPrefix string) string {
	u := NormalizeBotUsername(botUsername)
	if u == "" {
		return ""
	}
	if paramPrefix == "" {
		paramPrefix = DefaultReferralParamPrefix
	}
	return "https://t.me/" + u + "?startapp=" + paramPrefix
}

// ReferralLink builds a Mini App deep link: https://t.me/{bot}?startapp=ref_{telegram_id}
func ReferralLink(botUsername, referrerTelegramID, paramPrefix string) string {
	if referrerTelegramID == "" {
		return ""
	}
	base := ReferralLinkBase(botUsername, paramPrefix)
	if base == "" {
		return ""
	}
	return base + referrerTelegramID
}

// ParseReferralStartParam extracts referrer telegram_id from start_param / startapp payload.
// Supports "ref_777000" and plain "777000".
func ParseReferralStartParam(startParam, paramPrefix string) string {
	startParam = strings.TrimSpace(startParam)
	if startParam == "" {
		return ""
	}
	if paramPrefix == "" {
		paramPrefix = DefaultReferralParamPrefix
	}
	return strings.TrimPrefix(startParam, paramPrefix)
}
