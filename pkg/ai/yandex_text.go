package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/twelvepills-936/tgapp-/pkg/config"
)

const yandexCompletionURL = "https://llm.api.cloud.yandex.net/foundationModels/v1/completion"

func generateYandexText(
	ctx context.Context,
	cfg config.ConfigAI,
	messages []ChatMessage,
	_ string,
	modelSlug string,
	req TextRequest,
) (TextResponse, error) {
	modelURI := yandexModelURI(cfg.YandexFolderID, modelSlug)

	maxTokens := cfg.TextMaxOutputTokens
	if req.MaxTokens != nil && *req.MaxTokens > 0 {
		maxTokens = *req.MaxTokens
	}
	temperature := 0.6
	if req.Temperature != nil {
		temperature = *req.Temperature
	}

	yandexMessages := make([]map[string]string, 0, len(messages))
	for _, m := range messages {
		role := m.Role
		if role == "assistant" {
			role = "assistant"
		} else if role != "system" {
			role = "user"
		}
		yandexMessages = append(yandexMessages, map[string]string{
			"role": role,
			"text": m.Content,
		})
	}

	payload := map[string]any{
		"modelUri": modelURI,
		"completionOptions": map[string]any{
			"stream":      false,
			"temperature": temperature,
			"maxTokens":   maxTokens,
		},
		"messages": yandexMessages,
	}

	body, status, err := doJSON(ctx, cfg.HTTPTimeout, http.MethodPost, yandexCompletionURL, yandexHeaders(cfg.YandexAPIKey), payload)
	if err != nil {
		return TextResponse{}, err
	}
	if status != http.StatusOK {
		return TextResponse{}, &ProviderError{Provider: "yandex", Status: status, Message: truncate(string(body), 500)}
	}

	var resp struct {
		Result struct {
			Alternatives []struct {
				Message struct {
					Text string `json:"text"`
				} `json:"message"`
			} `json:"alternatives"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return TextResponse{}, fmt.Errorf("yandex: decode response: %w", err)
	}
	if len(resp.Result.Alternatives) == 0 || resp.Result.Alternatives[0].Message.Text == "" {
		return TextResponse{}, &ProviderError{Provider: "yandex", Message: "empty completion"}
	}

	return finalizeTextResponse(resp.Result.Alternatives[0].Message.Text, modelSlug), nil
}

func yandexModelURI(folderID, model string) string {
	model = strings.TrimSpace(model)
	if strings.HasPrefix(model, "gpt://") {
		return model
	}
	if model == "" {
		model = "yandexgpt/latest"
	}
	return fmt.Sprintf("gpt://%s/%s", folderID, model)
}

func yandexHeaders(apiKey string) map[string]string {
	return map[string]string{
		"Authorization": "Api-Key " + apiKey,
		"Content-Type":  "application/json",
	}
}

func doJSON(ctx context.Context, timeout time.Duration, method, url string, headers map[string]string, payload any) ([]byte, int, error) {
	var bodyReader io.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, 0, err
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, 0, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, &ProviderError{Provider: "http", Message: err.Error()}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return body, resp.StatusCode, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
