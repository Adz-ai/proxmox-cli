package tui

import "fmt"

// HumanBytes renders a byte count in binary units with one decimal place.
func HumanBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	value := float64(bytes)
	suffixes := []string{"KiB", "MiB", "GiB", "TiB", "PiB"}
	suffix := ""
	for _, s := range suffixes {
		value /= unit
		suffix = s
		if value < unit {
			break
		}
	}
	return fmt.Sprintf("%.1f %s", value, suffix)
}

// FormatUptime renders seconds as a compact duration such as "3d 4h" or "12m".
func FormatUptime(seconds uint64) string {
	if seconds == 0 {
		return "-"
	}
	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	minutes := (seconds % 3600) / 60
	switch {
	case days > 0:
		return fmt.Sprintf("%dd %dh", days, hours)
	case hours > 0:
		return fmt.Sprintf("%dh %dm", hours, minutes)
	case minutes > 0:
		return fmt.Sprintf("%dm", minutes)
	default:
		return fmt.Sprintf("%ds", seconds)
	}
}

// FormatPercent renders a 0..1 fraction as a percentage.
func FormatPercent(fraction float64) string {
	if fraction <= 0 {
		return "0.0%"
	}
	return fmt.Sprintf("%.1f%%", fraction*100)
}

// UsagePercent renders used/total as a percentage, or "-" when total is zero.
func UsagePercent(used, total uint64) string {
	if total == 0 {
		return "-"
	}
	return fmt.Sprintf("%.1f%%", float64(used)/float64(total)*100)
}

// truncate shortens a string to at most width cells, ending in an ellipsis.
func truncate(s string, width int) string {
	if width <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= width {
		return s
	}
	if width == 1 {
		return "…"
	}
	return string(runes[:width-1]) + "…"
}

// pad right-pads (or truncates) a string to exactly width cells.
func pad(s string, width int) string {
	s = truncate(s, width)
	for len([]rune(s)) < width {
		s += " "
	}
	return s
}
