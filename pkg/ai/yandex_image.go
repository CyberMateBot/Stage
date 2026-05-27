package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/twelvepills-936/tgapp-/pkg/config"
)

const (
	yandexImageAsyncURL = "https://ai.api.cloud.yandex.net/foundationModels/v1/imageGenerationAsync"
	yandexOperationsURL = "https://operation.api.cloud.yandex.net/operations"
)

func generateYandexImage(ctx context.Context, cfg config.ConfigAI, prompt string, req ImageRequest) (ImageResponse, error) {
	modelSlug := resolveImageModelSlug(req.Model, cfg.YandexImageModel)
	size := strings.TrimSpace(req.Size)
	if size == "" {
		size = cfg.YandexImageSize
	}

	modelURI := yandexArtModelURI(cfg.YandexFolderID, modelSlug)
	width, height := parseSizeRatio(size)

	payload := map[string]any{
		"modelUri": modelURI,
		"generationOptions": map[string]any{
			"aspectRatio": map[string]string{
				"widthRatio":  width,
				"heightRatio": height,
			},
		},
		"messages": []map[string]any{
			{"weight": "1", "text": prompt},
		},
	}

	body, status, err := doJSON(ctx, cfg.HTTPTimeout, http.MethodPost, yandexImageAsyncURL, yandexHeaders(cfg.YandexAPIKey), payload)
	if err != nil {
		return ImageResponse{}, err
	}
	if status != http.StatusOK {
		return ImageResponse{}, &ProviderError{Provider: "yandex", Status: status, Message: truncate(string(body), 500)}
	}

	var op struct {
		ID     string `json:"id"`
		Done   bool   `json:"done"`
		Error  *struct {
			Message string `json:"message"`
		} `json:"error"`
		Response *struct {
			Image string `json:"image"`
		} `json:"response"`
	}
	if err := json.Unmarshal(body, &op); err != nil {
		return ImageResponse{}, fmt.Errorf("yandex image: decode operation: %w", err)
	}
	if op.ID == "" {
		return ImageResponse{}, &ProviderError{Provider: "yandex", Message: "missing operation id"}
	}

	deadline := time.Now().Add(cfg.ImagePollTimeout)
	for time.Now().Before(deadline) {
		if op.Done {
			break
		}
		select {
		case <-ctx.Done():
			return ImageResponse{}, ctx.Err()
		case <-time.After(2 * time.Second):
		}

		pollBody, pollStatus, pollErr := doJSON(ctx, cfg.HTTPTimeout, http.MethodGet, yandexOperationsURL+"/"+op.ID, yandexHeaders(cfg.YandexAPIKey), nil)
		if pollErr != nil {
			return ImageResponse{}, pollErr
		}
		if pollStatus != http.StatusOK {
			return ImageResponse{}, &ProviderError{Provider: "yandex", Status: pollStatus, Message: truncate(string(pollBody), 500)}
		}
		if err := json.Unmarshal(pollBody, &op); err != nil {
			return ImageResponse{}, fmt.Errorf("yandex image: decode poll: %w", err)
		}
	}

	if op.Error != nil && op.Error.Message != "" {
		return ImageResponse{}, &ProviderError{Provider: "yandex", Message: op.Error.Message}
	}
	if op.Response == nil || op.Response.Image == "" {
		return ImageResponse{}, &ProviderError{Provider: "yandex", Message: "image generation timed out or empty response"}
	}

	return ImageResponse{
		ImageBase64: op.Response.Image,
		Model:       modelSlug,
	}, nil
}

func yandexArtModelURI(folderID, model string) string {
	model = strings.TrimSpace(model)
	if strings.HasPrefix(model, "art://") {
		return model
	}
	if strings.HasPrefix(model, "gpt://") {
		// some configs use gpt:// prefix for art models
		return strings.Replace(model, "gpt://", "art://", 1)
	}
	if model == "" {
		model = "yandex-art-2.0"
	}
	model = normalizeArtSlug(model)
	return fmt.Sprintf("art://%s/%s", folderID, model)
}

func parseSizeRatio(size string) (width, height string) {
	parts := strings.Split(strings.ToLower(strings.TrimSpace(size)), "x")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "1", "1"
}
