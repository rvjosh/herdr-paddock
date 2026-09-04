package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/neyham/herdr-paddock/internal/config"
	"github.com/neyham/herdr-paddock/internal/ui"
)

func main() {
	// Env first, flags on top: a shell profile sets the everyday shape of the
	// wall, a flag bends it for one run.
	env := config.FromEnv(os.Getenv)
	columns := flag.Int("columns", env.Columns, "cards per row (0 = auto, max 8)")
	previewLines := flag.Int("preview-lines", env.PreviewLines, "preview lines per card on a desktop (3-40)")
	noSummary := flag.Bool("no-summary", !env.Summary, "ignore the agent status block, always show the tail preview")
	theme := flag.String("theme", env.Theme, "palette: auto, dark or light")
	showVersion := flag.Bool("version", false, "print the version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Println("paddock " + ui.Version)
		return
	}
	cfg := config.Config{Columns: *columns, PreviewLines: *previewLines, Summary: !*noSummary, Theme: *theme}
	p := tea.NewProgram(ui.NewWithConfig(cfg), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "paddock: %v\n", err)
		os.Exit(1)
	}
}
