package ai

import "testing"

func TestResolveImageModelSlug(t *testing.T) {
	defaultSlug := "aliceai-image-art-3.0"

	tests := []struct {
		in   string
		want string
	}{
		{"alice-ai-art", "aliceai-image-art-3.0"},
		{"alice", "aliceai-image-art-3.0"},
		{"", "aliceai-image-art-3.0"},
		{"yandex-art-2.0", "yandex-art-2.0"},
		{"aliceai-image-art-3.0/latest", "aliceai-image-art-3.0"},
	}

	for _, tc := range tests {
		got := resolveImageModelSlug(tc.in, defaultSlug)
		if got != tc.want {
			t.Errorf("resolveImageModelSlug(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
