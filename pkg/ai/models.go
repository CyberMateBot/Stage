package ai

import "strings"

// TextModel describes a selectable text generation model for the frontend.
type TextModel struct {
	ID            string `json:"id"`
	Label         string `json:"label"`
	Group         string `json:"group"`
	Description   string `json:"description,omitempty"`
	Tier          string `json:"tier"` // fast | standard | pro
	Provider      string `json:"provider"` // yandex
	SupportsImage bool   `json:"supports_image,omitempty"`
}

type textModelDef struct {
	ID            string
	Label         string
	Group         string
	Description   string
	Tier          string
	Slug          string // Yandex model slug, e.g. gpt-oss-120b/latest
	UseResponses  bool
	UseOpenAIChat bool // OpenAI-compatible /v1/chat/completions only (e.g. DeepSeek)
	SupportsImage bool // multimodal input in chat (image attachment)
}

// textModelCatalog is the canonical list of Yandex text models exposed to the app.
var textModelCatalog = []textModelDef{
	{
		ID: "yandexgpt", Label: "YandexGPT", Group: "Yandex",
		Description: "Повседневные вопросы, письма и тексты на русском",
		Tier: "standard", Slug: "yandexgpt/latest",
	},
	{
		ID: "deepseek", Label: "DeepSeek V3.2", Group: "Yandex",
		Description: "Код, отладка, алгоритмы и пошаговые рассуждения",
		Tier: "pro", Slug: "deepseek-v32", UseOpenAIChat: true,
	},
	{
		ID: "gpt-oss-20b", Label: "GPT OSS 20B", Group: "Open-weight GPT",
		Description: "Черновики, короткие ответы и быстрые правки текста",
		Tier: "fast", Slug: "gpt-oss-20b/latest", UseResponses: true,
	},
	{
		ID: "gpt-oss-120b", Label: "GPT OSS 120B", Group: "Open-weight GPT",
		Description: "Сложные задачи, развёрнутые ответы и глубокие рассуждения",
		Tier: "pro", Slug: "gpt-oss-120b/latest", UseResponses: true,
	},
	{
		ID: "qwen3.6-35b", Label: "Qwen3.6 35B", Group: "Qwen",
		Description: "Точные ответы, структура и работа с длинным контекстом",
		Tier: "pro", Slug: "qwen3.6-35b-a3b", UseOpenAIChat: true, SupportsImage: true,
	},
	{
		ID: "qwen3-235b", Label: "Qwen3 235B", Group: "Qwen",
		Description: "Быстрые сводки и черновики по большим объёмам текста",
		Tier: "fast", Slug: "qwen3-235b-a22b-fp8/latest", UseResponses: true,
	},
}

// modelAliases maps client model ids to catalog ids.
var modelAliases = map[string]string{
	"yandex": "yandexgpt", "default": "yandexgpt",
	"gpt-oss-120b/latest": "gpt-oss-120b", "gpt_oss_120b": "gpt-oss-120b",
	"gpt-oss-20b/latest": "gpt-oss-20b", "gpt_oss_20b": "gpt-oss-20b",
	"qwen3.6-35b-a3b/latest": "qwen3.6-35b", "qwen3.6-35b": "qwen3.6-35b",
	"qwen3-235b-a22b-fp8/latest": "qwen3-235b", "qwen3-235b": "qwen3-235b",
	"deepseek-v32/latest": "deepseek",
}

func ListTextModels() []TextModel {
	out := make([]TextModel, 0, len(textModelCatalog))
	for _, m := range textModelCatalog {
		out = append(out, TextModel{
			ID: m.ID, Label: m.Label, Group: m.Group,
			Description: m.Description, Tier: m.Tier, Provider: "yandex",
			SupportsImage: m.SupportsImage,
		})
	}
	return out
}

func resolveTextModel(requested string, cfgSlug string) (def textModelDef, ok bool) {
	key := strings.ToLower(strings.TrimSpace(requested))
	if key == "" {
		key = "yandexgpt"
	}
	if id, found := modelAliases[key]; found {
		key = id
	}
	// direct slug pass-through: gpt://folder/slug or slug/latest
	key = strings.TrimPrefix(key, "gpt://")
	if i := strings.Index(key, "/"); i > 0 {
		key = key[strings.LastIndex(key, "/")+1:]
	}
	key = strings.TrimSuffix(key, "/latest")

	for _, m := range textModelCatalog {
		if m.ID == key || strings.TrimSuffix(m.Slug, "/latest") == key {
			return m, true
		}
	}
	// legacy: env default slug
	if cfgSlug != "" && (key == cfgSlug || strings.Contains(cfgSlug, key)) {
		return textModelDef{ID: key, Label: key, Slug: cfgSlug, Tier: "standard"}, true
	}
	return textModelDef{}, false
}
