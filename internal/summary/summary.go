// Package summary reads the little status block some agents leave at the end
// of their output:
//
//	STATUS: working
//	SUMMARY: wired the webhook route and verified signatures
//	NEXT: replay the failed invoice fixture
//	NEEDS ME: nothing right now
//
// It is plain text parsing — no model is asked what a pane means.
package summary

import "strings"

type Block struct {
	Status, Summary, Next, NeedsMe string
}

type field int

const (
	fStatus field = iota
	fSummary
	fNext
	fNeeds
)

var keyNames = []struct {
	name string
	f    field
}{
	{"STATUS", fStatus},
	{"SUMMARY", fSummary},
	{"NEXT", fNext},
	{"NEEDS ME", fNeeds},
	{"NEEDS_ME", fNeeds},
	{"NEEDS-ME", fNeeds},
	{"NEEDSME", fNeeds},
}

// noneWords are the ways an agent says "you're off the hook".
var noneWords = map[string]bool{
	"":                  true,
	"-":                 true,
	"—":                 true,
	"no":                true,
	"none":              true,
	"nothing":           true,
	"nothing right now": true,
	"n/a":               true,
	"nope":              true,
	"no action":         true,
}

// NeedsAttention reports whether the NEEDS ME field is asking for a human.
func (b Block) NeedsAttention() bool {
	v := strings.ToLower(strings.TrimSpace(b.NeedsMe))
	v = strings.TrimSpace(strings.TrimRight(v, ".!,;:·。！"))
	return !noneWords[v]
}

// ParseParagraphs takes cleaned transcript paragraphs, oldest first, and only
// accepts a block that is the agent's last word: a block buried above later
// prose is stale, not a status.
func ParseParagraphs(paras [][]string) (Block, bool) {
	if len(paras) == 0 {
		return Block{}, false
	}
	return Parse(paras[len(paras)-1])
}

// Parse reads one paragraph of already-cleaned lines, oldest first. The block
// starts at the last STATUS line; anything that is not a field or its
// continuation after that means the block is not the tail, so it is refused.
func Parse(lines []string) (Block, bool) {
	start := -1
	for i, ln := range lines {
		if f, _, ok := fieldAt(ln); ok && f == fStatus {
			start = i
		}
	}
	if start < 0 {
		return Block{}, false
	}
	var b Block
	var last field
	seen := map[field]bool{}
	open, ended := false, false
	for _, ln := range lines[start:] {
		if strings.TrimSpace(ln) == "" {
			ended, open = true, false
			continue
		}
		f, v, ok := fieldAt(ln)
		if !ok {
			if ended || !open {
				return Block{}, false // prose after the block: not a status tail
			}
			b.appendTo(last, strings.TrimSpace(ln))
			continue
		}
		ended = false
		*b.at(f) = v
		seen[f] = true
		last, open = f, true
	}
	if !seen[fSummary] && !seen[fNext] && !seen[fNeeds] {
		return Block{}, false // a lone STATUS says nothing worth a card
	}
	return b, true
}

func (b *Block) at(f field) *string {
	switch f {
	case fSummary:
		return &b.Summary
	case fNext:
		return &b.Next
	case fNeeds:
		return &b.NeedsMe
	default:
		return &b.Status
	}
}

func (b *Block) appendTo(f field, text string) {
	p := b.at(f)
	if *p == "" {
		*p = text
		return
	}
	*p += " " + text
}

// fieldAt reads a field key at the head of a line, past any bullet or
// markdown decoration. A "STATUS:" in the middle of a sentence is prose.
func fieldAt(line string) (field, string, bool) {
	s := stripDecor(line)
	for _, k := range keyNames {
		if len(s) < len(k.name) || !strings.EqualFold(s[:len(k.name)], k.name) {
			continue
		}
		rest := trimMark(s[len(k.name):])
		v, ok := afterSep(rest)
		if !ok {
			continue
		}
		return k.f, v, true
	}
	return fStatus, "", false
}

func stripDecor(s string) string {
	s = strings.TrimSpace(s)
	for {
		before := s
		for _, p := range []string{"**", "__", "`", "#", "- ", "* ", "• ", "> "} {
			if strings.HasPrefix(s, p) {
				s = strings.TrimSpace(s[len(p):])
				break
			}
		}
		if s == before {
			return s
		}
	}
}

// trimMark eats a closing bold/code marker, so "**STATUS**: x" and
// "**STATUS:** x" both read as a key.
func trimMark(s string) string {
	for {
		switch {
		case strings.HasPrefix(s, "**"):
			s = s[2:]
		case strings.HasPrefix(s, "__"):
			s = s[2:]
		case strings.HasPrefix(s, "`"):
			s = s[1:]
		default:
			return s
		}
	}
}

// afterSep splits the value off a key. A dash only separates when it is
// spaced, so "NEXT-gen work" stays prose.
func afterSep(rest string) (string, bool) {
	for _, sep := range []string{" — ", " – ", " - "} {
		if strings.HasPrefix(rest, sep) {
			return cleanValue(rest[len(sep):]), true
		}
	}
	t := strings.TrimLeft(rest, " \t")
	for _, sep := range []string{":", "=", "—", "–"} {
		if strings.HasPrefix(t, sep) {
			return cleanValue(t[len(sep):]), true
		}
	}
	return "", false
}

func cleanValue(v string) string {
	v = strings.TrimSpace(trimMark(strings.TrimSpace(v)))
	v = strings.TrimSpace(strings.TrimRight(v, "*_`"))
	return v
}
