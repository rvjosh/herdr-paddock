package config

import "testing"

func env(pairs map[string]string) func(string) string {
	return func(k string) string { return pairs[k] }
}

func TestDefaults(t *testing.T) {
	d := Default()
	if d.Columns != 0 || d.PreviewLines != 12 || !d.Summary {
		t.Fatalf("unexpected defaults: %+v", d)
	}
	if got := FromEnv(env(nil)); got != d {
		t.Fatalf("empty env should give the defaults, got %+v", got)
	}
	if got := FromEnv(nil); got != d {
		t.Fatalf("nil getenv should give the defaults, got %+v", got)
	}
}

func TestFromEnv(t *testing.T) {
	d := Default()
	cases := []struct {
		name string
		vars map[string]string
		want Config
	}{
		{"columns", map[string]string{"PADDOCK_COLUMNS": "4"}, Config{4, d.PreviewLines, true}},
		{"columns clamped high", map[string]string{"PADDOCK_COLUMNS": "99"}, Config{8, d.PreviewLines, true}},
		{"columns negative falls back", map[string]string{"PADDOCK_COLUMNS": "-2"}, d},
		{"columns junk falls back", map[string]string{"PADDOCK_COLUMNS": "wide"}, d},
		{"columns auto", map[string]string{"PADDOCK_COLUMNS": "0"}, d},
		{"preview lines", map[string]string{"PADDOCK_PREVIEW_LINES": "20"}, Config{0, 20, true}},
		{"preview lines clamped low", map[string]string{"PADDOCK_PREVIEW_LINES": "1"}, Config{0, 3, true}},
		{"preview lines clamped high", map[string]string{"PADDOCK_PREVIEW_LINES": "400"}, Config{0, 40, true}},
		{"preview lines zero falls back", map[string]string{"PADDOCK_PREVIEW_LINES": "0"}, d},
		{"preview lines junk falls back", map[string]string{"PADDOCK_PREVIEW_LINES": "lots"}, d},
		{"blank is unset", map[string]string{"PADDOCK_COLUMNS": "  ", "PADDOCK_SUMMARY": " "}, d},
		{"both", map[string]string{"PADDOCK_COLUMNS": "2", "PADDOCK_PREVIEW_LINES": "6"}, Config{2, 6, true}},
	}
	for _, c := range cases {
		if got := FromEnv(env(c.vars)); got != c.want {
			t.Fatalf("%s: FromEnv = %+v, want %+v", c.name, got, c.want)
		}
	}
}

func TestSummaryOff(t *testing.T) {
	for _, v := range []string{"0", "false", "FALSE", "no", "off", "Off"} {
		if FromEnv(env(map[string]string{"PADDOCK_SUMMARY": v})).Summary {
			t.Fatalf("PADDOCK_SUMMARY=%q should switch the status block off", v)
		}
	}
	for _, v := range []string{"1", "true", "yes", "on", "maybe"} {
		if !FromEnv(env(map[string]string{"PADDOCK_SUMMARY": v})).Summary {
			t.Fatalf("PADDOCK_SUMMARY=%q should leave the status block on", v)
		}
	}
}

func TestClamp(t *testing.T) {
	got := Config{Columns: -3, PreviewLines: -1, Summary: true}.Clamp()
	if got != Default() {
		t.Fatalf("nonsense values should fall back to the defaults, got %+v", got)
	}
	if got := (Config{Columns: 12, PreviewLines: 99}).Clamp(); got.Columns != 8 || got.PreviewLines != 40 {
		t.Fatalf("clamp should cap at the bounds, got %+v", got)
	}
}
