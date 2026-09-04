package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/neyham/herdr-paddock/internal/config"
)

// keepDark puts the dark palette back when the test is done: these are
// package vars, and every other test was written against the dark set.
func keepDark(t *testing.T) {
	t.Helper()
	t.Cleanup(func() { applyTheme(true) })
}

func fg(s lipgloss.Style) string {
	c, ok := s.GetForeground().(lipgloss.Color)
	if !ok {
		return ""
	}
	return string(c)
}

func bg(s lipgloss.Style) string {
	c, ok := s.GetBackground().(lipgloss.Color)
	if !ok {
		return ""
	}
	return string(c)
}

func TestApplyThemeSwapsTheBodyColour(t *testing.T) {
	keepDark(t)
	applyTheme(false)
	if got := fg(stBody); got != "#1f2335" {
		t.Fatalf("light body should be dark ink, got %q", got)
	}
	applyTheme(true)
	if got := fg(stBody); got != "#c0caf5" {
		t.Fatalf("dark body should be the old colour, got %q", got)
	}
}

func paletteColours(p palette) []string {
	return []string{
		p.text, p.body, p.dim, p.mute, p.wool, p.cyan, p.blue, p.good, p.warn,
		p.hot, p.bad, p.acc, p.line, p.faint, p.selB, p.badDim,
		p.chipBlck, p.chipWork, p.chipDone, p.chipIdle,
	}
}

func TestLightThemeCarriesNoDarkColours(t *testing.T) {
	keepDark(t)
	applyTheme(false)
	// Dark-only tones. The two palettes share one hex (#3b4261 is the dark
	// rule colour and the light muted text), so subtract what light uses.
	darkOnly := map[string]bool{}
	for _, c := range paletteColours(darkPalette) {
		darkOnly[c] = true
	}
	for _, c := range paletteColours(lightPalette) {
		delete(darkOnly, c)
	}
	styles := map[string]lipgloss.Style{
		"stText": stText, "stBody": stBody, "stDim": stDim, "stMute": stMute,
		"stWool": stWool, "stHead": stHead, "stWork": stWork, "stBlck": stBlck,
		"stDone": stDone, "stIdle": stIdle, "stErr": stErr, "stBar": stBar,
		"stSel": stSel, "stKey": stKey, "stLine": stLine, "stSelR": stSelR,
		"stBrand": stBrand, "stShadow": stShadow, "stNeedsLbl": stNeedsLbl,
		"stNeedsVal": stNeedsVal, "stChipBlck": stChipBlck, "stChipWork": stChipWork,
		"stChipDone": stChipDone, "stChipIdle": stChipIdle,
	}
	for name, st := range styles {
		if c := fg(st); darkOnly[c] {
			t.Fatalf("%s still paints with the dark colour %s", name, c)
		}
		if c := bg(st); c != "" && darkOnly[c] {
			t.Fatalf("%s still sits on the dark colour %s", name, c)
		}
	}
	// the bleat border tone is used raw, not through a style
	if darkOnly[string(cBadDim)] {
		t.Fatalf("the bleat low phase is still the dark tone")
	}
	// the chips keep their light tints, not the dark ones
	if bg(stChipBlck) != lightPalette.chipBlck || bg(stChipIdle) != lightPalette.chipIdle {
		t.Fatalf("chip backgrounds did not follow the theme: %q %q", bg(stChipBlck), bg(stChipIdle))
	}
}

func TestApplyThemeConfigPinsThePalette(t *testing.T) {
	keepDark(t)
	applyThemeConfig(config.ThemeLight)
	if fg(stBody) != lightPalette.body {
		t.Fatalf("light should pin the light body, got %q", fg(stBody))
	}
	applyThemeConfig(config.ThemeDark)
	if fg(stBody) != darkPalette.body {
		t.Fatalf("dark should pin the dark body, got %q", fg(stBody))
	}
}

func TestLightWallStillDraws(t *testing.T) {
	keepDark(t)
	m := sampleModel(100, 30)
	applyTheme(false) // after the model: NewWithConfig pins the theme itself
	plain := stripANSI(render(m))
	for _, need := range []string{"hooks", "pay", "╭", "paddock"} {
		if !strings.Contains(plain, need) {
			t.Fatalf("light wall lost %q:\n%s", need, plain)
		}
	}
}
