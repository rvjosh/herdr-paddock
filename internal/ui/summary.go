package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/neyham/herdr-paddock/internal/layout"
	"github.com/neyham/herdr-paddock/internal/model"
	"github.com/neyham/herdr-paddock/internal/summary"
)

type lineKind int

const (
	lnPlain lineKind = iota
	lnDim
	lnQuestion
	lnStatus
	lnNeeds
	lnSummary
	lnNext
)

type bodyLine struct {
	text string
	kind lineKind
}

// the little arrow between a field label and its value
const fieldArrow = "▸"

// cardBody is the one place a card's body is decided: cardHeight sizes the
// card from it and renderAgentCard paints it, so height and text can never
// disagree.
func cardBody(m Model, t model.TabRow, budget, maxLines int) []bodyLine {
	if maxLines < 1 {
		maxLines = 1
	}
	raw := cardPreview(m, t)
	if m.cfg.Summary {
		if b, ok := summary.ParseParagraphs(cleanParagraphs(raw)); ok {
			if body := blockBody(b, budget, maxLines); len(body) > 0 {
				return body
			}
		}
	}
	lines := clipTranscript(raw, budget, maxLines)
	_, status := tabKind(t)
	qs := -1
	if status == "blocked" {
		qs = questionStart(lines)
	}
	out := make([]bodyLine, 0, len(lines))
	for i, ln := range lines {
		kind := lnPlain
		switch {
		case qs >= 0 && i >= qs:
			kind = lnQuestion
		case qs >= 0:
			kind = lnDim
		}
		out = append(out, bodyLine{text: ln, kind: kind})
	}
	return out
}

// blockBody lays the status block out in the order a reader needs it: what
// the agent is doing, what it wants from you, what it did, what comes next.
func blockBody(b summary.Block, budget, maxLines int) []bodyLine {
	if budget < 4 {
		budget = 4
	}
	status := []string{layout.Truncate(fieldText("status", b.Status), budget)}
	var needs, sum, next []string
	if b.NeedsAttention() {
		needs = clip(layout.Wrap(fieldText("needs me", b.NeedsMe), budget), 3)
	}
	if b.Summary != "" {
		sum = layout.Wrap(fieldText("summary", b.Summary), budget)
	}
	if b.Next != "" {
		next = layout.Wrap(fieldText("next", b.Next), budget)
	}
	// Tight cards give up NEXT first, then SUMMARY, and keep the head of each
	// field rather than its tail: the start of a summary is the informative
	// part. The status line always survives.
	over := func() int { return len(status) + len(needs) + len(sum) + len(next) - maxLines }
	next = withValue(shrink(next, over(), 1))
	if over() > 0 {
		next = nil
	}
	sum = withValue(shrink(sum, over(), 1))
	needs = withValue(shrink(needs, over(), 1))
	out := make([]bodyLine, 0, maxLines)
	for _, group := range []struct {
		lines []string
		kind  lineKind
	}{{status, lnStatus}, {needs, lnNeeds}, {sum, lnSummary}, {next, lnNext}} {
		for _, ln := range group.lines {
			out = append(out, bodyLine{text: ln, kind: group.kind})
		}
	}
	if len(out) > maxLines {
		out = out[:maxLines]
	}
	return out
}

// shrink drops trailing lines of a field, never below floor.
func shrink(lines []string, over, floor int) []string {
	if over <= 0 || len(lines) <= floor {
		return lines
	}
	keep := len(lines) - over
	if keep < floor {
		keep = floor
	}
	return lines[:keep]
}

// withValue drops a field that got trimmed down to its label: on a narrow
// card the value can wrap past the lines the field kept, and a bare label
// says nothing.
func withValue(lines []string) []string {
	if len(lines) == 0 {
		return nil
	}
	joined := strings.Join(lines, " ")
	i := strings.Index(joined, fieldArrow)
	if i >= 0 && strings.TrimSpace(joined[i+len(fieldArrow):]) == "" {
		return nil
	}
	return lines
}

func fieldText(label, value string) string {
	return label + " " + fieldArrow + " " + value
}

// styleBody paints one body line. head marks the first line of a field, the
// only one carrying the label.
func styleBody(bl bodyLine, head bool) string {
	switch bl.kind {
	case lnDim:
		return stDim.Render(" " + bl.text)
	case lnQuestion:
		return stText.Render(" " + bl.text)
	case lnStatus:
		return labelled(bl.text, head, stDim, statusValueStyle(bl.text))
	case lnNeeds:
		return labelled(bl.text, head, stNeedsLbl, stNeedsVal)
	case lnSummary:
		return labelled(bl.text, head, stDim, stBody)
	case lnNext:
		return labelled(bl.text, head, stDim, stMute)
	default:
		return stBody.Render(" " + bl.text)
	}
}

func labelled(text string, head bool, label, value lipgloss.Style) string {
	i := strings.Index(text, fieldArrow)
	if !head || i < 0 {
		return value.Render(" " + text)
	}
	cut := i + len(fieldArrow)
	if cut < len(text) && text[cut] == ' ' {
		cut++
	}
	return label.Render(" "+text[:cut]) + value.Render(text[cut:])
}

// statusValueStyle colors the status word the way the card borders do, so a
// blocked agent reads red even inside its own summary.
func statusValueStyle(line string) lipgloss.Style {
	v := line
	if i := strings.Index(line, fieldArrow); i >= 0 {
		v = line[i+len(fieldArrow):]
	}
	v = strings.ToLower(v)
	col := cText
	switch {
	case hasAny(v, "blocked", "waiting", "needs", "stuck"):
		col = cBad
	case hasAny(v, "done", "complete", "finished", "shipped"):
		col = cBlue
	case hasAny(v, "working", "running", "in progress"):
		col = cWarn
	}
	return lipgloss.NewStyle().Foreground(col).Bold(true)
}

func hasAny(s string, words ...string) bool {
	for _, w := range words {
		if strings.Contains(s, w) {
			return true
		}
	}
	return false
}
