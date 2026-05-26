package ai

import (
	"context"
	"strings"

	"github.com/twelvepills-936/tgapp-/pkg/config"
)

// TextRequest is the body for POST /v1/generate/text.
type TextRequest struct {
	Prompt      string        `json:"prompt"`
	Text        string        `json:"text"` // alias for prompt
	Model       string        `json:"model"`
	System      string        `json:"system"`
	Messages    []ChatMessage `json:"messages"`
	Temperature *float64      `json:"temperature"`
	MaxTokens   *int          `json:"max_tokens"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Text    string `json:"text"` // some clients send "text" instead of "content"
}

// TextResponse is returned by POST /v1/generate/text.
type TextResponse struct {
	Text   string `json:"text"`
	Model  string `json:"model"`
	Format string `json:"format"` // markdown — render with Markdown + LaTeX ($...$) on frontend
}

// ModelsResponse is returned by GET /v1/generate/models.
type ModelsResponse struct {
	TextModels []TextModel `json:"text_models"`
}

// ImageRequest is the body for POST /v1/generate/image.
type ImageRequest struct {
	Prompt      string `json:"prompt"`
	Text        string `json:"text"`
	Model       string `json:"model"`
	Size        string `json:"size"`
	AspectRatio string `json:"aspect_ratio"`
}

// ImageResponse is returned by POST /v1/generate/image.
type ImageResponse struct {
	ImageURL    string `json:"image_url,omitempty"`
	ImageBase64 string `json:"image_base64,omitempty"`
	Model       string `json:"model"`
}

// Service routes generation to configured providers.
type Service struct {
	cfg config.ConfigAI
}

func NewService(cfg config.ConfigAI) *Service {
	return &Service{cfg: cfg}
}

func (s *Service) ListModels() ModelsResponse {
	return ModelsResponse{TextModels: ListTextModels()}
}

func (s *Service) GenerateText(ctx context.Context, req TextRequest) (TextResponse, error) {
	prompt, messages, err := normalizeTextInput(req)
	if err != nil {
		return TextResponse{}, err
	}

	provider := strings.ToLower(strings.TrimSpace(req.Model))
	switch provider {
	case "openai", "gpt":
		if s.cfg.OpenAITextEnabled() {
			return generateOpenAICompatText(ctx, s.cfg, "https://api.openai.com/v1", s.cfg.OpenAITextModel, messages, prompt, req)
		}
	case "gemini", "wavespeed":
		if s.cfg.WavespeedTextEnabled() {
			return generateOpenAICompatText(ctx, s.cfg, s.cfg.GeminiAPIBaseURL, s.cfg.GeminiModel, messages, prompt, req)
		}
	}

	if s.cfg.YandexTextEnabled() {
		def, ok := resolveTextModel(req.Model, s.cfg.YandexGPTModel)
		if !ok {
			// unknown slug — try completion API with raw model name
			return generateYandexText(ctx, s.cfg, messages, prompt, req.Model, req)
		}
		if def.UseResponses {
			return generateYandexResponsesText(ctx, s.cfg, messages, def.Slug, req)
		}
		slug := def.Slug
		if slug == "" {
			slug = s.cfg.YandexGPTModel
		}
		return generateYandexText(ctx, s.cfg, messages, prompt, slug, req)
	}

	if s.cfg.OpenAITextEnabled() {
		return generateOpenAICompatText(ctx, s.cfg, "https://api.openai.com/v1", s.cfg.OpenAITextModel, messages, prompt, req)
	}
	if s.cfg.WavespeedTextEnabled() {
		return generateOpenAICompatText(ctx, s.cfg, s.cfg.GeminiAPIBaseURL, s.cfg.GeminiModel, messages, prompt, req)
	}

	return TextResponse{}, ErrNotConfigured
}

func (s *Service) GenerateImage(ctx context.Context, req ImageRequest) (ImageResponse, error) {
	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		prompt = strings.TrimSpace(req.Text)
	}
	if prompt == "" {
		return ImageResponse{}, ErrPromptEmpty
	}

	provider := strings.ToLower(strings.TrimSpace(req.Model))
	switch provider {
	case "", "yandex", "alice", "yandex-art", "default":
		if s.cfg.YandexTextEnabled() {
			return generateYandexImage(ctx, s.cfg, prompt, req)
		}
	case "nano-banana", "wavespeed", "banana":
		if s.cfg.WavespeedImageEnabled() {
			return generateWavespeedImage(ctx, s.cfg, prompt, req)
		}
	default:
		if s.cfg.YandexTextEnabled() {
			return generateYandexImage(ctx, s.cfg, prompt, req)
		}
	}

	if s.cfg.YandexTextEnabled() {
		return generateYandexImage(ctx, s.cfg, prompt, req)
	}
	if s.cfg.WavespeedImageEnabled() {
		return generateWavespeedImage(ctx, s.cfg, prompt, req)
	}

	return ImageResponse{}, ErrNotConfigured
}

func normalizeTextInput(req TextRequest) (prompt string, messages []ChatMessage, err error) {
	prompt = strings.TrimSpace(req.Prompt)
	if prompt == "" {
		prompt = strings.TrimSpace(req.Text)
	}

	messages = req.Messages
	if len(messages) == 0 && prompt != "" {
		messages = []ChatMessage{{Role: "user", Content: prompt}}
	}
	if len(messages) == 0 {
		return "", nil, ErrPromptEmpty
	}

	if strings.TrimSpace(req.System) != "" {
		messages = append([]ChatMessage{{Role: "system", Content: req.System}}, messages...)
	}

	if len(messages) > 0 && messages[0].Role == "system" {
		messages[0].Content = mergeInstructions(messages[0].Content)
	} else {
		messages = append([]ChatMessage{{Role: "system", Content: mergeInstructions("")}}, messages...)
	}

	for i := range messages {
		if messages[i].Content == "" {
			messages[i].Content = messages[i].Text
		}
		if messages[i].Role == "" {
			messages[i].Role = "user"
		}
	}
	return prompt, messages, nil
}
