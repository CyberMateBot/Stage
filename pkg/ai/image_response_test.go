package ai

import (
	"strings"
	"testing"
)

func TestFinalizeImageResponse_setsDataURL(t *testing.T) {
	out := FinalizeImageResponse(ImageResponse{
		ImageBase64: "iVBORw0KGgo=",
		Model:       "alice-ai-art",
	})
	if out.ImageURL == "" {
		t.Fatal("expected image_url data URL")
	}
	if !strings.HasPrefix(out.ImageURL, "data:image/png;base64,") {
		t.Fatalf("unexpected url: %q", out.ImageURL)
	}
}
