package tui

import "testing"

func TestHumanBytes(t *testing.T) {
	cases := []struct {
		in   uint64
		want string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KiB"},
		{1536, "1.5 KiB"},
		{8 * 1024 * 1024 * 1024, "8.0 GiB"},
		{3 * 1024 * 1024 * 1024 * 1024, "3.0 TiB"},
	}
	for _, tc := range cases {
		if got := HumanBytes(tc.in); got != tc.want {
			t.Errorf("HumanBytes(%d) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestFormatUptime(t *testing.T) {
	cases := []struct {
		in   uint64
		want string
	}{
		{0, "-"},
		{45, "45s"},
		{300, "5m"},
		{3900, "1h 5m"},
		{90000, "1d 1h"},
	}
	for _, tc := range cases {
		if got := FormatUptime(tc.in); got != tc.want {
			t.Errorf("FormatUptime(%d) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestFormatPercent(t *testing.T) {
	if got := FormatPercent(0.1234); got != "12.3%" {
		t.Errorf("FormatPercent(0.1234) = %q", got)
	}
	if got := FormatPercent(0); got != "0.0%" {
		t.Errorf("FormatPercent(0) = %q", got)
	}
}

func TestUsagePercent(t *testing.T) {
	if got := UsagePercent(512, 1024); got != "50.0%" {
		t.Errorf("UsagePercent(512, 1024) = %q", got)
	}
	if got := UsagePercent(1, 0); got != "-" {
		t.Errorf("UsagePercent with zero total = %q", got)
	}
}

func TestPadAndTruncate(t *testing.T) {
	if got := pad("ab", 4); got != "ab  " {
		t.Errorf("pad(ab, 4) = %q", got)
	}
	if got := pad("abcdef", 4); got != "abc…" {
		t.Errorf("pad(abcdef, 4) = %q", got)
	}
	if got := truncate("abc", 0); got != "" {
		t.Errorf("truncate to zero width = %q", got)
	}
}
