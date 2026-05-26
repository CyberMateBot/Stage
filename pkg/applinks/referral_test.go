package applinks

import "testing"

func TestReferralLink(t *testing.T) {
	t.Parallel()

	tests := []struct {
		bot      string
		telegram string
		want     string
	}{
		{"CyberMate_bot", "12345", "https://t.me/CyberMate_bot?start=12345"},
		{"@CyberMate_bot", "999", "https://t.me/CyberMate_bot?start=999"},
		{"", "123", ""},
		{"CyberMate_bot", "", ""},
	}

	for _, tc := range tests {
		if got := ReferralLink(tc.bot, tc.telegram); got != tc.want {
			t.Fatalf("ReferralLink(%q, %q) = %q, want %q", tc.bot, tc.telegram, got, tc.want)
		}
	}
}
