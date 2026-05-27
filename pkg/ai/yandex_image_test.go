package ai

import "testing"

func TestYandexArtModelURI(t *testing.T) {
	const folder = "b1g6rv1n9e4qau0o3qi8"

	tests := []struct {
		in   string
		want string
	}{
		{"aliceai-image-art-3.0", "art://" + folder + "/aliceai-image-art-3.0"},
		{"aliceai-image-art-3.0/latest", "art://" + folder + "/aliceai-image-art-3.0"},
		{"art://aliceai-image-art-3.0/latest", "art://" + folder + "/aliceai-image-art-3.0"},
		{"art://" + folder + "/aliceai-image-art-3.0", "art://" + folder + "/aliceai-image-art-3.0"},
	}

	for _, tc := range tests {
		got := yandexArtModelURI(folder, tc.in)
		if got != tc.want {
			t.Errorf("yandexArtModelURI(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
