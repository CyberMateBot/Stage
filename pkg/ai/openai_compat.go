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

	oaiMessages := make([]map[string]any, 0, len(messages)+1)
	for _, m := range messages {
		oaiMessages = append(oaiMessages, map[string]any{
			"role":    m.Role,
			"content": m.Content,
		})
	}

	if strings.TrimSpace(req.ImageBase64) != "" {
		mime := strings.TrimSpace(req.ImageMimeType)
		if mime == "" {
			mime = "image/jpeg"
		}
		dataURL := "data:" + mime + ";base64," + strings.TrimSpace(req.ImageBase64)

		// Attach the image to the last user message.
		lastUser := -1
		for i := len(oaiMessages) - 1; i >= 0; i-- {
			if role, ok := oaiMessages[i]["role"].(string); ok && role == "user" {
				lastUser = i
				break
			}
		}
		if lastUser >= 0 {
			prev := oaiMessages[lastUser]
			text := ""
			if s, ok := prev["content"].(string); ok {
				text = s
			}
			prev["content"] = []map[string]any{
				{"type": "text", "text": text},
				{"type": "image_url", "image_url": map[string]any{"url": dataURL}},
			}
			oaiMessages[lastUser] = prev
		} else {
			oaiMessages = append(oaiMessages, map[string]any{
				"role": "user",
				"content": []map[string]any{
					{"type": "text", "text": strings.TrimSpace(req.Prompt)},
					{"type": "image_url", "image_url": map[string]any{"url": dataURL}},
				},
			})
		}
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

	return finalizeTextResponse(resp.Choices[0].Message.Content, model), nil
}

func modelProviderName(baseURL string) string {
	if strings.Contains(baseURL, "wavespeed") {
		return "wavespeed"
	}
	return "openai"
}
