package summary

import (
	"strings"
	"testing"
)

func lines(s string) []string {
	return strings.Split(strings.TrimSpace(s), "\n")
}

func TestParseBlocks(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want Block
	}{
		{
			name: "canonical",
			in: `STATUS: working
SUMMARY: Wired the Stripe webhook route and verified signatures
NEXT: replay invoice.payment_failed, then run the full suite
NEEDS ME: nothing right now`,
			want: Block{
				Status:  "working",
				Summary: "Wired the Stripe webhook route and verified signatures",
				Next:    "replay invoice.payment_failed, then run the full suite",
				NeedsMe: "nothing right now",
			},
		},
		{
			name: "markdown bold",
			in: `**STATUS:** working
**SUMMARY**: shipped the parser
- **NEXT** — run the suite`,
			want: Block{Status: "working", Summary: "shipped the parser", Next: "run the suite"},
		},
		{
			name: "lowercase keys and bullets",
			in: `* status: blocked
* summary: waiting on the provider choice
* needs me: pick paddle or stripe`,
			want: Block{Status: "blocked", Summary: "waiting on the provider choice", NeedsMe: "pick paddle or stripe"},
		},
		{
			name: "needs me spellings",
			in: `STATUS: done
NEEDS_ME: sign the release`,
			want: Block{Status: "done", NeedsMe: "sign the release"},
		},
		{
			name: "needs-me dashed",
			in: `STATUS: done
NEEDS-ME = sign the release
NEXT: tag v2`,
			want: Block{Status: "done", NeedsMe: "sign the release", Next: "tag v2"},
		},
		{
			name: "needsme joined",
			in: `STATUS: done
NEEDSME: sign the release`,
			want: Block{Status: "done", NeedsMe: "sign the release"},
		},
		{
			name: "continuation lines",
			in: `STATUS: working
SUMMARY: Wired the Stripe webhook route and verified signatures
against the fixture events.
NEXT: replay the failed invoice
and then the full suite`,
			want: Block{
				Status:  "working",
				Summary: "Wired the Stripe webhook route and verified signatures against the fixture events.",
				Next:    "replay the failed invoice and then the full suite",
			},
		},
		{
			name: "last status wins",
			in: `STATUS: stale
SUMMARY: an older block
STATUS: working
SUMMARY: the current one`,
			want: Block{Status: "working", Summary: "the current one"},
		},
	}
	for _, c := range cases {
		got, ok := Parse(lines(c.in))
		if !ok {
			t.Fatalf("%s: block not recognised", c.name)
		}
		if got != c.want {
			t.Fatalf("%s: got %+v, want %+v", c.name, got, c.want)
		}
	}
}

func TestParseRefuses(t *testing.T) {
	cases := []struct{ name, in string }{
		{"no status", "SUMMARY: shipped it\nNEXT: rest"},
		{"status alone", "STATUS: working"},
		{"prose after the block", "STATUS: working\nSUMMARY: shipped it\n\nthen I noticed a failing test."},
		{"mid-sentence key", "I set STATUS: working in the header row.\nSUMMARY: not a block either"},
		{"plain prose", "wired the webhook and moved on"},
		{"empty", ""},
	}
	for _, c := range cases {
		if b, ok := Parse(lines(c.in)); ok {
			t.Fatalf("%s: should not parse, got %+v", c.name, b)
		}
	}
}

func TestParseParagraphsWantsTheLastWord(t *testing.T) {
	block := lines(`STATUS: working
SUMMARY: shipped it`)
	prose := []string{"one more thing I noticed"}
	if _, ok := ParseParagraphs([][]string{prose, block}); !ok {
		t.Fatal("a block in the last paragraph should parse")
	}
	if _, ok := ParseParagraphs([][]string{block, prose}); ok {
		t.Fatal("a block above later prose is stale, not a status")
	}
	if _, ok := ParseParagraphs(nil); ok {
		t.Fatal("no paragraphs, no block")
	}
}

func TestNeedsAttention(t *testing.T) {
	cases := map[string]bool{
		"":                    false,
		"-":                   false,
		"—":                   false,
		"no":                  false,
		"None":                false,
		"nothing":             false,
		"nothing right now":   false,
		"Nothing right now.":  false,
		"n/a":                 false,
		"nope":                false,
		"no action":           false,
		"approve the openapi": true,
		"pick paddle":         true,
		"nothing works, help": true,
	}
	for v, want := range cases {
		if got := (Block{NeedsMe: v}).NeedsAttention(); got != want {
			t.Fatalf("NeedsAttention(%q) = %v, want %v", v, got, want)
		}
	}
}
