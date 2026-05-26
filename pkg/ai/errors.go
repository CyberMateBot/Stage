package ai

import (
	"errors"
	"strconv"
)

var (
	ErrNotConfigured = errors.New("ai provider is not configured")
	ErrPromptEmpty   = errors.New("prompt is required")
)

// ProviderError is returned when an upstream AI API fails.
type ProviderError struct {
	Provider string
	Status   int
	Message  string
}

func (e *ProviderError) Error() string {
	if e.Status > 0 {
		return e.Provider + " API error (" + strconv.Itoa(e.Status) + "): " + e.Message
	}
	return e.Provider + " API error: " + e.Message
}
