package ai

import (
	"strings"
	"testing"
)

func TestFormatModelText_UnicodeStars(t *testing.T) {
	in := "2.∗∗Упрощение∗∗"
	out := FormatModelText(in)
	if out != "2.**Упрощение**" {
		t.Fatalf("got %q", out)
	}
}

func TestFormatModelText_WrapBareLatex(t *testing.T) {
	in := `D_f=\mathbb{R}\setminus\{2\}.`
	out := FormatModelText(in)
	if !strings.Contains(out, "$") {
		t.Fatalf("expected $ wrapped latex, got %q", out)
	}
}
