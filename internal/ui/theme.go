package ui

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/neyham/herdr-paddock/internal/config"
)

// palette is one terminal's worth of colour. Two exist: the Tokyo Night set
// the wall has always shipped, and a Tokyo Night Day-ish set for light
// terminals, where the dark body colour was all but invisible.
type palette struct {
	text, body, dim, mute, wool            string
	cyan, blue, good, warn, hot, bad, acc  string
	line, faint, selB, badDim              string
	chipBlck, chipWork, chipDone, chipIdle string
}

// Tokyo Night-ish, at home next to a typical dark terminal setup.
var darkPalette = palette{
	text: "#c0caf5", body: "#c0caf5", dim: "#565f89", mute: "#a9b1d6", wool: "#c4b5a0",
	cyan: "#7dcfff", blue: "#7aa2f7", good: "#9ece6a", warn: "#e0af68", hot: "#ff9e64",
	bad: "#f7768e", acc: "#bb9af7",
	line: "#3b4261", faint: "#24283b", selB: "#292e42", badDim: "#7a3b4a",
	chipBlck: "#402734", chipWork: "#3d3223", chipDone: "#24304a", chipIdle: "#2a3627",
}

// Tokyo Night Day-ish. The body colour is genuinely dark rather than a light
// grey: on white and off-white terminals that is the whole point.
var lightPalette = palette{
	text: "#1f2335", body: "#1f2335", dim: "#6c7086", mute: "#3b4261", wool: "#8a7a5a",
	cyan: "#007197", blue: "#2e7de9", good: "#587539", warn: "#8c6c3e", hot: "#b15c00",
	bad: "#f52a65", acc: "#7847bd",
	line: "#a8aecb", faint: "#d5d6db", selB: "#e9e9ec", badDim: "#e8a0b0",
	chipBlck: "#fbdde4", chipWork: "#f2e8d2", chipDone: "#dbe6fb", chipIdle: "#e2ecd5",
}

// Colours and the styles built from them. Vars, not consts: applyTheme
// repaints the whole set in one go, so nothing else has to know a theme
// exists. NO_COLOR is lipgloss's business, not ours.
var (
	cText  lipgloss.Color
	cBody  lipgloss.Color
	cDim   lipgloss.Color
	cMute  lipgloss.Color
	cWool  lipgloss.Color
	cCyan  lipgloss.Color
	cBlue  lipgloss.Color
	cGood  lipgloss.Color
	cWarn  lipgloss.Color
	cHot   lipgloss.Color
	cBad   lipgloss.Color
	cAcc   lipgloss.Color
	cLine  lipgloss.Color
	cFaint lipgloss.Color
	cSelB  lipgloss.Color
	// low phase of the blocked "bleat" breathing border
	cBadDim lipgloss.Color

	stText lipgloss.Style
	stBody lipgloss.Style
	stDim  lipgloss.Style
	stMute lipgloss.Style
	stWool lipgloss.Style
	stHead lipgloss.Style
	// Match herdr sidebar dots: red blocked, yellow working, blue done, green idle.
	stWork  lipgloss.Style
	stBlck  lipgloss.Style
	stDone  lipgloss.Style
	stIdle  lipgloss.Style
	stErr   lipgloss.Style
	stBar   lipgloss.Style
	stSel   lipgloss.Style
	stKey   lipgloss.Style
	stLine  lipgloss.Style
	stSelR  lipgloss.Style
	stBrand lipgloss.Style
	// soft drop shadow under the selected card: half blocks in a near-bg tone
	stShadow lipgloss.Style
	// the one line on a card that is addressed to you: the agent's NEEDS ME
	stNeedsLbl lipgloss.Style
	stNeedsVal lipgloss.Style
	// title chips: status-tinted background, xiaohongshu cover feel
	stChipBlck lipgloss.Style
	stChipWork lipgloss.Style
	stChipDone lipgloss.Style
	stChipIdle lipgloss.Style
)

func init() { applyTheme(true) }

// applyTheme paints the package palette. Called once at startup, before
// bubbletea owns the screen, and from tests.
func applyTheme(dark bool) {
	p := lightPalette
	if dark {
		p = darkPalette
	}
	cText = lipgloss.Color(p.text)
	cBody = lipgloss.Color(p.body)
	cDim = lipgloss.Color(p.dim)
	cMute = lipgloss.Color(p.mute)
	cWool = lipgloss.Color(p.wool)
	cCyan = lipgloss.Color(p.cyan)
	cBlue = lipgloss.Color(p.blue)
	cGood = lipgloss.Color(p.good)
	cWarn = lipgloss.Color(p.warn)
	cHot = lipgloss.Color(p.hot)
	cBad = lipgloss.Color(p.bad)
	cAcc = lipgloss.Color(p.acc)
	cLine = lipgloss.Color(p.line)
	cFaint = lipgloss.Color(p.faint)
	cSelB = lipgloss.Color(p.selB)
	cBadDim = lipgloss.Color(p.badDim)

	stText = lipgloss.NewStyle().Foreground(cText)
	stBody = lipgloss.NewStyle().Foreground(cBody)
	stDim = lipgloss.NewStyle().Foreground(cDim)
	stMute = lipgloss.NewStyle().Foreground(cMute)
	stWool = lipgloss.NewStyle().Foreground(cWool)
	stHead = lipgloss.NewStyle().Foreground(cBlue)
	stWork = lipgloss.NewStyle().Foreground(cWarn).Bold(true)
	stBlck = lipgloss.NewStyle().Foreground(cBad).Bold(true)
	stDone = lipgloss.NewStyle().Foreground(cBlue)
	stIdle = lipgloss.NewStyle().Foreground(cGood)
	stErr = lipgloss.NewStyle().Foreground(cBad)
	stBar = lipgloss.NewStyle().Foreground(cDim)
	stSel = lipgloss.NewStyle().Foreground(cCyan).Bold(true)
	stKey = lipgloss.NewStyle().Foreground(cCyan)
	stLine = lipgloss.NewStyle().Foreground(cLine)
	stSelR = lipgloss.NewStyle().Foreground(cText).Background(cSelB)
	stBrand = lipgloss.NewStyle().Foreground(cCyan).Bold(true)
	stShadow = lipgloss.NewStyle().Foreground(cFaint)
	stNeedsLbl = lipgloss.NewStyle().Foreground(cHot).Bold(true)
	stNeedsVal = lipgloss.NewStyle().Foreground(cText).Bold(true)
	stChipBlck = lipgloss.NewStyle().Foreground(cBad).Background(lipgloss.Color(p.chipBlck)).Bold(true)
	stChipWork = lipgloss.NewStyle().Foreground(cWarn).Background(lipgloss.Color(p.chipWork))
	stChipDone = lipgloss.NewStyle().Foreground(cBlue).Background(lipgloss.Color(p.chipDone))
	stChipIdle = lipgloss.NewStyle().Foreground(cGood).Background(lipgloss.Color(p.chipIdle))
}

// applyThemeConfig turns the configured theme into a palette. "auto" asks the
// terminal for its background — once, from here, before bubbletea takes the
// screen. A terminal that won't answer gets the dark set.
func applyThemeConfig(theme string) {
	switch theme {
	case config.ThemeLight:
		applyTheme(false)
	case config.ThemeDark:
		applyTheme(true)
	default:
		applyTheme(lipgloss.HasDarkBackground())
	}
}
