package ai

import "strings"

// resolveImageModelSlug maps frontend model ids to Yandex ART slugs (without art:// prefix).
func resolveImageModelSlug(requested string, defaultSlug string) string {
	key := strings.ToLower(strings.TrimSpace(requested))
	switch key {
	case "", "yandex", "yandex-art", "default":
		return normalizeArtSlug(defaultSlug)
	case "alice", "alice-ai-art", "alice-ai-art-3", "aliceai-image-art-3.0":
		return "aliceai-image-art-3.0"
	case "yandex-art-2", "yandex-art-2.0":
		return "yandex-art-2.0"
	default:
		return normalizeArtSlug(requested)
	}
}

func normalizeArtSlug(slug string) string {
	slug = strings.TrimSpace(slug)
	slug = strings.TrimPrefix(slug, "art://")
	slug = strings.TrimPrefix(slug, "gpt://")
	slug = strings.TrimSuffix(slug, "/latest")
	if slug == "" {
		return ""
	}
	// art://<folder>/<model> or <folder>/<model> → model is the last segment
	if strings.Contains(slug, "/") {
		parts := strings.Split(slug, "/")
		slug = parts[len(parts)-1]
		if slug == "latest" && len(parts) > 1 {
			slug = parts[len(parts)-2]
		}
	}
	return slug
}
