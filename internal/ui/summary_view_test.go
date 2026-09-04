package ui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/neyham/herdr-paddock/internal/config"
	"github.com/neyham/herdr-paddock/internal/layout"
	"github.com/neyham/herdr-paddock/internal/model"
	"github.com/neyham/herdr-paddock/internal/summary"
)

// an agent that ends its output with the status block, chrome and all
const blockPreview = "Wired the endpoint and verified the signatures.\n\n" +
	"STATUS: working\n" +
	"SUMMARY: wired the Stripe webhook route against the fixture events\n" +
	"NEXT: replay the failed invoice, then the full suite\n" +
	"NEEDS ME: approve the regenerated openapi.json before I publish\n"

func blockModel(w, h int, cfg config.Config, preview string) Model {
	m := NewWithConfig(cfg)
	m.width, m.height = w, h
	m.snap.Tabs = []model.TabRow{
		{TabID: "w1:t1", PaneID: "p1", WSLabel: "shop", TabLabel: "webhooks", Agent: "codex", Status: "working", Title: "Wire Stripe webhooks"},
	}
	m.previews = map[string]string{"p1": preview}
	m.selID = "w1:t1"
	return m
}

func parsedBlock(t *testing.T) summary.Block {
	t.Helper()
	b, ok := summary.ParseParagraphs(cleanParagraphs(blockPreview))
	if !ok {
		t.Fatal("the fixture block should parse")
	}
	return b
}

func TestFeedColumnCount(t *testing.T) {
	d := config.Default()
	four := config.Config{Columns: 4, PreviewLines: 12, Summary: true}
	six := config.Config{Columns: 6, PreviewLines: 12, Summary: true}
	cases := []struct {
		name       string
		cfg        config.Config
		w, n, want int
	}{
		{"desktop default is three", d, 100, 6, 3},
		{"wide desktop stays three", d, 140, 6, 3},
		{"desktop with two agents", d, 100, 2, 2},
		{"explicit four", four, 140, 6, 4},
		{"explicit six capped by the 18-cell rule", six, 80, 6, 4},
		{"explicit four on a phone caps too", four, 40, 6, 2},
		{"explicit never below one", six, 20, 6, 1},
		{"phone keeps two", d, 40, 2, 2},
		{"narrow drops to one", d, 30, 2, 1},
		{"one agent is one column", d, 140, 1, 1},
	}
	for _, c := range cases {
		if got := feedCols(c.cfg, c.w, c.n); got != c.want {
			t.Fatalf("%s: feedCols(%d cols, w=%d, n=%d) = %d, want %d", c.name, c.cfg.Columns, c.w, c.n, got, c.want)
		}
	}
}

// A configured column count must never squeeze a card below 18 cells.
func TestConfiguredColumnsKeepCardsReadable(t *testing.T) {
	m := blockModel(80, 30, config.Config{Columns: 6, PreviewLines: 12, Summary: true}, "line one")
	m.snap.Tabs = nil
	for i := 0; i < 6; i++ {
		id := fmt.Sprintf("t%d", i)
		m.snap.Tabs = append(m.snap.Tabs, model.TabRow{TabID: id, PaneID: "p" + id, WSLabel: "infra", TabLabel: id, Agent: "pi", Status: "idle"})
		m.previews["p"+id] = "line one"
	}
	m.selID = "t0"
	slots, _, _ := glanceSlots(m, 80, glanceFeedH(30))
	for _, s := range slots {
		if s.w < minCardW {
			t.Fatalf("card %d is %d cells wide, below the %d minimum", s.idx, s.w, minCardW)
		}
	}
}

func TestMaxBodyForBands(t *testing.T) {
	d := config.Default()
	ten := config.Config{Columns: 0, PreviewLines: 10, Summary: true}
	cases := []struct {
		band layout.Band
		cfg  config.Config
		want int
	}{
		{layout.Phone, d, 5},
		{layout.Compact, d, 8},
		{layout.Full, d, 12},
		{layout.Phone, ten, 5},
		{layout.Compact, ten, 8},
		{layout.Full, ten, 10},
	}
	for _, c := range cases {
		if got := maxBodyFor(c.band, c.cfg); got != c.want {
			t.Fatalf("maxBodyFor(%s, %d preview lines) = %d, want %d", c.band, c.cfg.PreviewLines, got, c.want)
		}
	}
}

func TestPreviewReadLines(t *testing.T) {
	if got := previewReadLines(config.Default()); got != 36 {
		t.Fatalf("default preview read should cover 12 lines of card, got %d", got)
	}
	if got := previewReadLines(config.Config{PreviewLines: 5}); got != 28 {
		t.Fatalf("short previews should keep the 28-line floor, got %d", got)
	}
}

func TestStatusBlockOnCard(t *testing.T) {
	m := blockModel(100, 30, config.Default(), blockPreview)
	plain := stripANSI(render(m))
	for _, need := range []string{"status ▸", "working", "needs me ▸", "openapi.json", "summary ▸", "Stripe webhook route", "next ▸", "the full suite"} {
		if !strings.Contains(plain, need) {
			t.Fatalf("missing %q:\n%s", need, plain)
		}
	}
	for _, bad := range []string{"STATUS:", "SUMMARY:", "NEEDS ME:", "NEXT:"} {
		if strings.Contains(plain, bad) {
			t.Fatalf("raw key %q should not reach the card:\n%s", bad, plain)
		}
	}
	// what the reader must act on sits right under the status line
	if strings.Index(plain, "needs me ▸") > strings.Index(plain, "summary ▸") {
		t.Fatalf("needs me should come before summary:\n%s", plain)
	}
}

func TestStatusBlockWithoutNeedsMe(t *testing.T) {
	preview := strings.Replace(blockPreview, "NEEDS ME: approve the regenerated openapi.json before I publish", "NEEDS ME: nothing right now", 1)
	plain := stripANSI(render(blockModel(100, 30, config.Default(), preview)))
	if strings.Contains(plain, "needs me") {
		t.Fatalf("a block that needs nothing should not feature a needs me line:\n%s", plain)
	}
	if !strings.Contains(plain, "summary ▸") {
		t.Fatalf("the rest of the block should still show:\n%s", plain)
	}
}

func TestSummaryOffFallsBackToTailPreview(t *testing.T) {
	cfg := config.Default()
	cfg.Summary = false
	plain := stripANSI(render(blockModel(100, 30, cfg, blockPreview)))
	if !strings.Contains(plain, "SUMMARY:") {
		t.Fatalf("with summary off the card should show the raw tail:\n%s", plain)
	}
	if strings.Contains(plain, "status ▸") {
		t.Fatalf("with summary off nothing should be parsed:\n%s", plain)
	}
}

// The block is the agent's last word, but its own input box and footer sit
// under it in the pane. The cleaner drops those, so the block still parses.
func TestStatusBlockSurvivesAgentChrome(t *testing.T) {
	preview := blockPreview +
		"\n" +
		"  ╭───────────────────── 2026-08-08｜Checkout｜Webhooks ──────────╮\n" +
		"  │ ❯                                                            │\n" +
		"  ╰────────────────────── Grok 4.6 (xhigh) · always-approve ─────╯\n" +
		"  Shift+Tab:mode  │  Ctrl+.:shortcuts\n"
	plain := stripANSI(render(blockModel(100, 30, config.Default(), preview)))
	for _, need := range []string{"status ▸", "needs me ▸", "summary ▸"} {
		if !strings.Contains(plain, need) {
			t.Fatalf("missing %q under agent chrome:\n%s", need, plain)
		}
	}
	if strings.Contains(plain, "always-approve") {
		t.Fatalf("chrome leaked onto the card:\n%s", plain)
	}
}

func TestBlockCardsFitTheirWidth(t *testing.T) {
	for _, w := range []int{100, 40} {
		m := blockModel(w, 30, config.Default(), blockPreview)
		for i := 0; i < 3; i++ {
			id := fmt.Sprintf("t%d", i)
			m.snap.Tabs = append(m.snap.Tabs, model.TabRow{TabID: id, PaneID: "p" + id, WSLabel: "infra", TabLabel: id, Agent: "pi", Status: "working"})
			m.previews["p"+id] = blockPreview
		}
		for n, line := range strings.Split(render(m), "\n") {
			if got := layout.Width(stripANSI(line)); got > w {
				t.Fatalf("w=%d line %d width %d: %q", w, n, got, stripANSI(line))
			}
		}
	}
}

func TestBlockCardHeightMatchesItsBody(t *testing.T) {
	m := blockModel(100, 30, config.Default(), blockPreview)
	viewH := glanceFeedH(30)
	slots, _, feed := glanceSlots(m, 100, viewH)
	if len(slots) != 1 {
		t.Fatalf("want one card, got %d", len(slots))
	}
	innerW := slots[0].w - 2
	body := cardBody(m, feed[0], innerW-1, maxBodyFor(layout.BandOf(100), m.cfg))
	if len(body) != 4 {
		t.Fatalf("want the four block fields, got %d: %+v", len(body), body)
	}
	if slots[0].h != 3+len(body) {
		t.Fatalf("card height %d should be 3 + %d body lines", slots[0].h, len(body))
	}
}

// cardHeight sizes a card from cardBody and renderAgentCard fills it from
// cardBody again, with the smaller budget the chosen height allows. If the
// second answer were shorter the card would grow a blank row.
func TestCardHeightAgreesWithWhatIsPainted(t *testing.T) {
	for _, w := range []int{40, 60, 100} {
		m := blockModel(w, 24, config.Default(), blockPreview)
		m.previews["p2"] = strings.Replace(blockPreview, "NEEDS ME: approve the regenerated openapi.json before I publish", "NEEDS ME: nothing right now", 1)
		m.previews["p3"] = "a plain agent that just keeps talking about the retry queue and its backoff"
		m.snap.Tabs = append(m.snap.Tabs,
			model.TabRow{TabID: "w1:t2", PaneID: "p2", WSLabel: "docs", TabLabel: "api", Agent: "claude", Status: "done"},
			model.TabRow{TabID: "w1:t3", PaneID: "p3", WSLabel: "infra", TabLabel: "cron", Agent: "pi", Status: "idle"},
		)
		slots, _, feed := glanceSlots(m, w, glanceFeedH(24))
		maxBody := maxBodyFor(layout.BandOf(w), m.cfg)
		for _, s := range slots {
			innerW := max(8, s.w-2)
			budget := innerW - 1
			if budget < 4 {
				budget = innerW
			}
			bodyH := max(1, s.h-3)
			sized := len(cardBody(m, feed[s.idx], budget, maxBody))
			painted := len(cardBody(m, feed[s.idx], budget, bodyH))
			if want := min(sized, bodyH); painted != want {
				t.Fatalf("w=%d card %d: sized %d lines, painted %d, want %d", w, s.idx, sized, painted, want)
			}
		}
	}
}

// A tight card gives up NEXT first, then trims SUMMARY, and always keeps the
// status line and the head of each field.
func TestBlockBodyTrimOrder(t *testing.T) {
	b := parsedBlock(t)
	full := blockBody(b, 40, 12)
	if len(full) < 5 {
		t.Fatalf("wide card should show the whole block, got %d lines", len(full))
	}
	for _, limit := range []int{6, 4, 3, 2, 1} {
		got := blockBody(b, 40, limit)
		if len(got) > limit {
			t.Fatalf("maxLines=%d produced %d lines", limit, len(got))
		}
		if got[0].kind != lnStatus {
			t.Fatalf("maxLines=%d dropped the status line: %+v", limit, got)
		}
		kinds := map[lineKind]int{}
		for _, ln := range got {
			kinds[ln.kind]++
		}
		if limit <= 3 && kinds[lnNext] > 0 {
			t.Fatalf("maxLines=%d should have dropped next: %+v", limit, got)
		}
		if limit == 1 && len(got) != 1 {
			t.Fatalf("one line means status only: %+v", got)
		}
	}
	// trimming keeps the head of a field, not its tail
	short := blockBody(b, 40, 4)
	if !strings.HasPrefix(short[len(short)-1].text, "summary ▸") && short[len(short)-1].kind == lnSummary {
		t.Fatalf("a trimmed summary should keep its first line: %+v", short)
	}
}
