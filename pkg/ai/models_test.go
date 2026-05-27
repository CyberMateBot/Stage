package ai

import "testing"

func TestResolveTextModel(t *testing.T) {
	def, ok := resolveTextModel("gpt-oss-120b", "")
	if !ok || !def.UseResponses || def.Slug != "gpt-oss-120b/latest" {
		t.Fatalf("gpt-oss-120b: %+v ok=%v", def, ok)
	}

	def, ok = resolveTextModel("deepseek", "")
	if !ok || !def.UseOpenAIChat || def.Slug != "deepseek-v32" {
		t.Fatalf("deepseek: %+v ok=%v", def, ok)
	}

	def, ok = resolveTextModel("yandexgpt", "")
	if !ok || def.UseResponses {
		t.Fatalf("yandexgpt: %+v", def)
	}

	def, ok = resolveTextModel("qwen3-235b-a22b-fp8/latest", "")
	if !ok || def.ID != "qwen3-235b" {
		t.Fatalf("qwen alias: %+v", def)
	}
}

func TestListTextModels(t *testing.T) {
	models := ListTextModels()
	if len(models) < 6 {
		t.Fatalf("expected at least 6 models, got %d", len(models))
	}
}
