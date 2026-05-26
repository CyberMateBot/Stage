package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/twelvepills-936/tgapp-/pkg/config"
)

func generateOpenAICompatText(
	ctx context.Context,
	cfg config.ConfigAI,
	baseURL string,
	model string,
	messages []ChatMessage,
	_ string,
	req TextRequest,
) (TextResponse, error) {
	baseURL = strings.TrimRight(baseURL, "/")
	url := baseURL + "/chat/completions"

	oaiMessages := make([]map[string]string, 0, len(messages))
	for _, m := range messages {
		oaiMessages = append(oaiMessages, map[string]string{
			"role":    m.Role,
			"content": m.Content,
		})
	}

	maxTokens := cfg.TextMaxOutputTokens
	if req.MaxTokens != nil && *req.MaxTokens > 0 {
		maxTokens = *req.MaxTokens
	}

	payload := map[string]any{
		"model":    model,
		"messages": oaiMessages,
		"max_tokens": maxTokens,
	}
	if req.Temperature != nil {
		payload["temperature"] = *req.Temperature
	}

	apiKey := cfg.OpenAIAPIKey
	if strings.Contains(baseURL, "wavespeed") {
		apiKey = cfg.WavespeedAPIKey
	}
	headers := map[string]string{
		"Authorization": "Bearer " + apiKey,
		"Content-Type":  "application/json",
	}

	body, status, err := doJSON(ctx, cfg.HTTPTimeout, http.MethodPost, url, headers, payload)
	if err != nil {
		return TextResponse{}, err
	}
	if status != http.StatusOK {
		return TextResponse{}, &ProviderError{Provider: modelProviderName(baseURL), Status: status, Message: truncate(string(body), 500)}
	}

	var resp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return TextResponse{}, fmt.Errorf("openai compat: decode: %w", err)
	}
	if len(resp.Choices) == 0 {
		return TextResponse{}, &ProviderError{Provider: modelProviderName(baseURL), Message: "empty completion"}
	}

	return TextResponse{
		Text:  resp.Choices[0].Message.Content,
		Model: model,
	}, nil
}

func modelProviderName(baseURL string) string {
	if strings.Contains(baseURL, "wavespeed") {
		return "wavespeed"
	}
	return "openai"
}
