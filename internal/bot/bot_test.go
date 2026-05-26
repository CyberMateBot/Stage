package bot

import (
	"strings"
	"testing"
)

func TestMiniAppOpenURL(t *testing.T) {
	t.Setenv("TELEGRAM_MINI_APP_URL", "")
	t.Setenv("TELEGRAM_BOT_USERNAME", "CyberMate_bot")

	got := miniAppOpenURL()
	want := "https://t.me/CyberMate_bot?startapp"
	if got != want {
		t.Fatalf("miniAppOpenURL() = %q, want %q", got, want)
	}

	t.Setenv("TELEGRAM_MINI_APP_URL", "https://app.example.com")
	if miniAppOpenURL() != "https://app.example.com" {
		t.Fatalf("custom TELEGRAM_MINI_APP_URL not used")
	}
}

func TestStartWelcomeText(t *testing.T) {
	if startWelcomeText == "" {
		t.Fatal("startWelcomeText is empty")
	}
	if strings.Contains(startWelcomeText, "реклам") || strings.Contains(startWelcomeText, "Добро пожаловать") {
		t.Fatalf("old welcome text still present: %q", startWelcomeText)
	}
}
