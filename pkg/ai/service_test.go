package ai

import (
	"testing"
)

func TestNormalizeTextInput(t *testing.T) {
	_, msgs, err := normalizeTextInput(TextRequest{Prompt: "hi"})
	if err != nil || len(msgs) != 1 || msgs[0].Content != "hi" {
		t.Fatalf("unexpected: msgs=%v err=%v", msgs, err)
	}

	_, _, err = normalizeTextInput(TextRequest{})
	if err != ErrPromptEmpty {
		t.Fatalf("expected ErrPromptEmpty, got %v", err)
	}
}
