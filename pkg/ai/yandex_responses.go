package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/twelvepills-936/tgapp-/pkg/config"
)

const yandexResponsesURL = "https://ai.api.cloud.yandex.net/v1/responses"

type yandexResponsesRequest struct {
	Model           string  `json:"model"`
	Temperature     float64 `json:"temperature"`
	Instructions    string  `json:"instructions"`
	Input           string  `json:"input"`
	MaxOutputTokens int     `json:"max_output_tokens"`
}

type yandexResponsesData struct {
	Output []struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	} `json:"output"`
}

func (r yandexResponsesData) outputText() string {
	for _, output := range r.Output {
		if len(output.Content) > 0 && output.Content[0].Text != "" {
			return output.Content[0].Text
		}
	}
	return ""
}

func generateYandexResponsesText(
	ctx context.Context,
	cfg config.ConfigAI,
	messages []ChatMessage,
	modelSlug string,
	req TextRequest,
) (TextResponse, error) {
	instructions, input := messagesToResponsesPayload(messages)
	instructions = mergeInstructions(instructions)

	maxTokens := cfg.TextMaxOutputTokens
	if req.MaxTokens != nil && *req.MaxTokens > 0 {
		maxTokens = *req.MaxTokens
	}
	temperature := 0.3
	if req.Temperature != nil {
		temperature = *req.Temperature
	}

	payload := yandexResponsesRequest{
		Model:           yandexModelURI(cfg.YandexFolderID, modelSlug),
		Temperature:     temperature,
		Instructions:    instructions,
		Input:           input,
		MaxOutputTokens: maxTokens,
	}

	headers := yandexHeaders(cfg.YandexAPIKey)
	headers["OpenAI-Project"] = cfg.YandexFolderID

	body, status, err := doJSON(ctx, cfg.HTTPTimeout, http.MethodPost, yandexResponsesURL, headers, payload)
	if err != nil {
		return TextResponse{}, err
	}
	if status != http.StatusOK {
		return TextResponse{}, &ProviderError{Provider: "yandex-responses", Status: status, Message: truncate(string(body), 500)}
	}

	var resp yandexResponsesData
	if err := json.Unmarshal(body, &resp); err != nil {
		return TextResponse{}, fmt.Errorf("yandex responses: decode: %w", err)
	}
	text := resp.outputText()
	if text == "" {
		return TextResponse{}, &ProviderError{Provider: "yandex-responses", Message: "empty completion"}
	}

	return finalizeTextResponse(text, modelSlug), nil
}

func messagesToResponsesPayload(messages []ChatMessage) (instructions, input string) {
	var systemParts []string
	var userParts []string
	for _, m := range messages {
		switch m.Role {
		case "system":
			systemParts = append(systemParts, m.Content)
		case "assistant":
			userParts = append(userParts, "Assistant: "+m.Content)
		default:
			userParts = append(userParts, m.Content)
		}
	}
	instructions = strings.Join(systemParts, "\n")
	input = strings.Join(userParts, "\n")
	if input == "" && len(userParts) == 0 && len(messages) > 0 {
		input = messages[len(messages)-1].Content
	}
	return instructions, input
}
