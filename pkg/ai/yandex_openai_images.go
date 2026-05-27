package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/twelvepills-936/tgapp-/pkg/config"
)

const yandexOpenAIImagesURL = "https://ai.api.cloud.yandex.net/v1/images/generations"

// generateYandexOpenAIImage uses Yandex AI Studio OpenAI-compatible Images API.
// This path is required for Alice AI ART in some setups.
func generateYandexOpenAIImage(ctx context.Context, cfg config.ConfigAI, prompt string, req ImageRequest) (ImageResponse, error) {
	folderID := strings.TrimSpace(cfg.YandexFolderID)
	if folderID == "" {
		return ImageResponse{}, &ProviderError{Provider: "yandex-images", Message: "YANDEX_GPT_FOLDER_ID is not configured"}
	}

	// Build model URI exactly like AI Studio examples:
	// art://{folder}/{aliceai-image-art-3.0/latest}
	model := strings.TrimSpace(cfg.YandexImageModel)
	if model == "" {
		model = "aliceai-image-art-3.0/latest"
	}
	// Allow passing a full URI via env.
	if strings.HasPrefix(model, "art://") {
		// ok
	} else {
		model = "art://" + folderID + "/" + strings.TrimPrefix(model, "/")
	}

	size := strings.TrimSpace(req.Size)
	if size == "" {
		size = strings.TrimSpace(cfg.YandexImageSize)
	}
	if size == "" {
		size = "1024x1024"
	}

	payload := map[string]any{
		"model":           model,
		"prompt":          prompt,
		"size":            size,
		"response_format": "b64_json",
	}

	// Yandex OpenAI-compatible APIs use Api-Key + OpenAI-Project (folder id).
	headers := yandexHeaders(cfg.YandexAPIKey)
	headers["OpenAI-Project"] = folderID

	body, status, err := doJSON(ctx, cfg.HTTPTimeout, http.MethodPost, yandexOpenAIImagesURL, headers, payload)
	if err != nil {
		return ImageResponse{}, err
	}
	if status != http.StatusOK {
		return ImageResponse{}, &ProviderError{Provider: "yandex-images", Status: status, Message: truncate(string(body), 500)}
	}

	var resp struct {
		Data []struct {
			B64JSON string `json:"b64_json"`
			URL     string `json:"url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return ImageResponse{}, fmt.Errorf("yandex images: decode: %w", err)
	}
	if len(resp.Data) == 0 {
		return ImageResponse{}, &ProviderError{Provider: "yandex-images", Message: "empty image"}
	}

	item := resp.Data[0]
	b64 := strings.TrimSpace(item.B64JSON)
	url := strings.TrimSpace(item.URL)
	if b64 == "" && url == "" {
		return ImageResponse{}, &ProviderError{Provider: "yandex-images", Message: "empty image"}
	}

	out := ImageResponse{
		ImageBase64: b64,
		ImageURL:    url,
		Model:       strings.TrimSpace(req.Model),
	}
	return FinalizeImageResponse(out), nil
}

