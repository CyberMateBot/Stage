package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/twelvepills-936/tgapp-/pkg/config"
)

const (
	// YandexART ImageGenerationAsync REST endpoint
	yandexImageAsyncURL = "https://llm.api.cloud.yandex.net/foundationModels/v1/imageGenerationAsync"
	yandexOperationsURL = "https://operation.api.cloud.yandex.net/operations"
)

func generateYandexImage(ctx context.Context, cfg config.ConfigAI, prompt string, req ImageRequest) (ImageResponse, error) {
	folderID := strings.TrimSpace(cfg.YandexFolderID)
	if folderID == "" {
		return ImageResponse{}, &ProviderError{Provider: "yandex", Message: "YANDEX_GPT_FOLDER_ID is not configured"}
	}

	requestedModel := strings.ToLower(strings.TrimSpace(req.Model))
	modelSlug := resolveImageModelSlug(req.Model, cfg.YandexImageModel)
	size := strings.TrimSpace(req.Size)
	if size == "" {
		size = cfg.YandexImageSize
	}

	doRequest := func(slug string) ([]byte, int, error) {
		modelURI := yandexArtModelURI(folderID, slug)
		slog.InfoContext(ctx, "yandex image request",
			slog.String("requested_model", strings.TrimSpace(req.Model)),
			slog.String("resolved_slug", slug),
			slog.String("model_uri", modelURI),
			slog.String("size", size),
		)
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
		return doJSON(ctx, cfg.HTTPTimeout, http.MethodPost, yandexImageAsyncURL, yandexHeaders(cfg.YandexAPIKey), payload)
	}

	body, status, err := doRequest(modelSlug)
	if err != nil {
		return ImageResponse{}, err
	}
	if status != http.StatusOK {
		msg := truncate(string(body), 500)
	// For Alice AI ART the OpenAI-compatible Images API is used (see generateYandexOpenAIImage).
	// Keep this fallback only for other YandexART slugs.
	if status == http.StatusBadRequest &&
		(requestedModel == "alice" || requestedModel == "alice-ai-art") &&
		strings.Contains(msg, "not allowed") &&
		strings.Contains(msg, "aliceai-image-art-3.0") {
			slog.WarnContext(ctx, "alice ai art not allowed, falling back to yandex-art-2.0")
			body, status, err = doRequest("yandex-art-2.0")
			if err != nil {
				return ImageResponse{}, err
			}
			if status != http.StatusOK {
				return ImageResponse{}, &ProviderError{Provider: "yandex", Status: status, Message: truncate(string(body), 500)}
			}
		} else {
			return ImageResponse{}, &ProviderError{Provider: "yandex", Status: status, Message: msg}
		}
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

	return FinalizeImageResponse(ImageResponse{
		ImageBase64: op.Response.Image,
		Model:       modelSlug,
	}), nil
}

func yandexArtModelURI(folderID, model string) string {
	model = normalizeArtSlug(model)
	if model == "" {
		model = "yandex-art-2.0"
	}
	return fmt.Sprintf("art://%s/%s", strings.TrimSpace(folderID), model)
}

func parseSizeRatio(size string) (width, height string) {
	parts := strings.Split(strings.ToLower(strings.TrimSpace(size)), "x")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "1", "1"
}
