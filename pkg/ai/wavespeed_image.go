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

const wavespeedAPIBase = "https://api.wavespeed.ai/api/v3"

func generateWavespeedImage(ctx context.Context, cfg config.ConfigAI, prompt string, req ImageRequest) (ImageResponse, error) {
	modelPath := strings.Trim(cfg.NanoBananaModel, "/")
	if !strings.Contains(modelPath, "/") {
		modelPath = "google/" + modelPath
	}
	submitURL := wavespeedAPIBase + "/" + modelPath
	if !strings.HasSuffix(submitURL, "text-to-image") && !strings.HasSuffix(submitURL, "edit") {
		submitURL += "/text-to-image"
	}

	payload := map[string]any{
		"prompt":               prompt,
		"resolution":           cfg.NanoBananaResolution,
		"output_format":        cfg.NanoBananaOutputFmt,
		"enable_sync_mode":     cfg.NanoBananaSyncMode,
		"enable_base64_output": cfg.NanoBananaBase64Out,
	}
	if ar := strings.TrimSpace(req.AspectRatio); ar != "" {
		payload["aspect_ratio"] = ar
	}

	headers := map[string]string{
		"Authorization": "Bearer " + cfg.WavespeedAPIKey,
		"Content-Type":  "application/json",
	}

	body, status, err := doJSON(ctx, cfg.HTTPTimeout, http.MethodPost, submitURL, headers, payload)
	if err != nil {
		return ImageResponse{}, err
	}
	if status != http.StatusOK && status != http.StatusCreated {
		return ImageResponse{}, &ProviderError{Provider: "wavespeed", Status: status, Message: truncate(string(body), 500)}
	}

	var submit struct {
		Data struct {
			ID     string `json:"id"`
			Status string `json:"status"`
			Outputs []string `json:"outputs"`
		} `json:"data"`
		ID string `json:"id"`
	}
	_ = json.Unmarshal(body, &submit)

	requestID := submit.Data.ID
	if requestID == "" {
		requestID = submit.ID
	}
	if requestID == "" {
		// sync mode may return outputs immediately
		if len(submit.Data.Outputs) > 0 {
			return ImageResponse{ImageURL: submit.Data.Outputs[0], Model: modelPath}, nil
		}
		return ImageResponse{}, &ProviderError{Provider: "wavespeed", Message: "missing prediction id"}
	}

	if cfg.NanoBananaSyncMode && len(submit.Data.Outputs) > 0 {
		return ImageResponse{ImageURL: submit.Data.Outputs[0], Model: modelPath}, nil
	}

	deadline := time.Now().Add(cfg.ImagePollTimeout)
	pollInterval := time.Duration(cfg.NanoBananaPollMS) * time.Millisecond

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ImageResponse{}, ctx.Err()
		case <-time.After(pollInterval):
		}

		resultURL := wavespeedAPIBase + "/predictions/" + requestID + "/result"
		pollBody, pollStatus, pollErr := doJSON(ctx, cfg.HTTPTimeout, http.MethodGet, resultURL, headers, nil)
		if pollErr != nil {
			return ImageResponse{}, pollErr
		}
		if pollStatus != http.StatusOK {
			continue
		}

		var result struct {
			Data struct {
				Status  string   `json:"status"`
				Outputs []string `json:"outputs"`
				Output  string   `json:"output"`
			} `json:"data"`
			Status  string   `json:"status"`
			Outputs []string `json:"outputs"`
		}
		if err := json.Unmarshal(pollBody, &result); err != nil {
			return ImageResponse{}, fmt.Errorf("wavespeed: decode poll: %w", err)
		}

		status := result.Data.Status
		if status == "" {
			status = result.Status
		}
		outputs := result.Data.Outputs
		if len(outputs) == 0 {
			outputs = result.Outputs
		}
		if result.Data.Output != "" {
			outputs = append(outputs, result.Data.Output)
		}

		switch strings.ToLower(status) {
		case "completed", "succeeded", "success":
			if len(outputs) == 0 {
				return ImageResponse{}, &ProviderError{Provider: "wavespeed", Message: "completed without output url"}
			}
			return ImageResponse{ImageURL: outputs[0], Model: modelPath}, nil
		case "failed", "error":
			return ImageResponse{}, &ProviderError{Provider: "wavespeed", Message: "generation failed: " + string(pollBody)}
		}
	}

	return ImageResponse{}, &ProviderError{Provider: "wavespeed", Message: "image generation timed out"}
}
