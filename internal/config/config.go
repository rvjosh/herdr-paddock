// Package config holds the few knobs that change how the wall is laid out.
// Everything here is plain data: the UI reads it, nothing writes it back.
package config

import (
	"strconv"
	"strings"
)

const (
	maxColumns      = 8
	minPreviewLines = 3
	maxPreviewLines = 40
)

type Config struct {
	Columns      int  // 0 = auto
	PreviewLines int  // max preview body lines per card on a desktop (Full band, width >= 80)
	Summary      bool // detect the agent status block
}

func Default() Config {
	return Config{Columns: 0, PreviewLines: 12, Summary: true}
}

// Clamp keeps a config inside the range the wall can actually draw. Values
// that make no sense at all (negative, zero preview) fall back to the
// default for that field rather than to the nearest bound.
func (c Config) Clamp() Config {
	d := Default()
	if c.Columns < 0 {
		c.Columns = d.Columns
	}
	if c.Columns > maxColumns {
		c.Columns = maxColumns
	}
	switch {
	case c.PreviewLines <= 0:
		c.PreviewLines = d.PreviewLines
	case c.PreviewLines < minPreviewLines:
		c.PreviewLines = minPreviewLines
	case c.PreviewLines > maxPreviewLines:
		c.PreviewLines = maxPreviewLines
	}
	return c
}

// FromEnv reads PADDOCK_COLUMNS, PADDOCK_PREVIEW_LINES and PADDOCK_SUMMARY.
// A value that isn't a number keeps the default for its field: a typo in a
// shell profile should never leave someone with an unusable wall.
func FromEnv(getenv func(string) string) Config {
	cfg := Default()
	if getenv == nil {
		return cfg
	}
	if n, ok := envInt(getenv("PADDOCK_COLUMNS")); ok {
		cfg.Columns = n
	}
	if n, ok := envInt(getenv("PADDOCK_PREVIEW_LINES")); ok {
		cfg.PreviewLines = n
	}
	if v := strings.TrimSpace(getenv("PADDOCK_SUMMARY")); v != "" {
		switch strings.ToLower(v) {
		case "0", "false", "no", "off":
			cfg.Summary = false
		}
	}
	return cfg.Clamp()
}

func envInt(raw string) (int, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, false
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0, false
	}
	return n, true
}
