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
	if strings.HasPrefix(slug, "art://") {
		rest := strings.TrimPrefix(slug, "art://")
		if i := strings.Index(rest, "/"); i >= 0 {
			slug = rest[i+1:]
		} else {
			slug = rest
		}
	}
	return strings.TrimSuffix(slug, "/latest")
}
