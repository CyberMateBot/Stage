package ai

import (
	"regexp"
	"strings"
)

// Default instructions so models return Telegram/Mini App friendly Markdown + LaTeX.
const defaultTextFormattingInstructions = `Формат ответа для мобильного приложения:
- Используй обычный Markdown: **жирный**, списки через "- ", заголовки ##.
- Математику пиши только в LaTeX: inline $...$, блочные формулы $$...$$.
- Не выводи сырой LaTeX без $ (например не пиши \\frac{x}{y} отдельно — пиши $\\frac{x}{y}$).
- Не разбивай звёздочки ** на отдельные строки; пиши **Упрощение** в одной строке.
- Таблицы — в Markdown (| col |), без HTML.
- Ответ на русском, если пользователь пишет по-русски.`

var (
	unicodeStarReplacer = strings.NewReplacer(
		"∗", "*",  // U+2217
		"＊", "*",  // U+FF0A
		"⁎", "*",  // U+204E
		"✱", "*",
		"﹡", "*",
	)
	latexCmdRe = regexp.MustCompile(`\\[a-zA-Z]+`)
)

// FormatModelText normalizes model output for Markdown/LaTeX renderers on the frontend.
func FormatModelText(text string) string {
	text = unicodeStarReplacer.Replace(text)
	text = collapseBrokenBoldMarkers(text)
	text = wrapBareLatexSegments(text)
	text = strings.TrimSpace(text)
	return text
}

func mergeInstructions(userSystem string) string {
	parts := []string{defaultTextFormattingInstructions}
	if strings.TrimSpace(userSystem) != "" {
		parts = append(parts, strings.TrimSpace(userSystem))
	}
	return strings.Join(parts, "\n\n")
}

// collapseBrokenBoldMarkers fixes "** word **" and "∗∗word∗∗" spacing issues.
func collapseBrokenBoldMarkers(s string) string {
	s = unicodeStarReplacer.Replace(s)
	// Remove spaces inside bold markers: * * word * * -> **word**
	for strings.Contains(s, "* *") {
		s = strings.ReplaceAll(s, "* *", "**")
	}
	// Newlines between lone asterisks (model sometimes emits vertical **)
	lines := strings.Split(s, "\n")
	var b strings.Builder
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "*" || trimmed == "**" {
			b.WriteString("**")
			continue
		}
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(line)
	}
	return b.String()
}

// wrapBareLatexSegments wraps fragments containing LaTeX commands but not $ delimiters.
func wrapBareLatexSegments(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if strings.Contains(line, "$") || !latexCmdRe.MatchString(line) {
			continue
		}
		// Whole line looks like a formula
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") && !strings.HasPrefix(trimmed, "|") {
			lines[i] = "$" + trimmed + "$"
		}
	}
	return strings.Join(lines, "\n")
}

func finalizeTextResponse(text, model string) TextResponse {
	return TextResponse{
		Text:   FormatModelText(text),
		Model:  model,
		Format: "markdown",
	}
}
