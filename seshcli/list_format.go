package seshcli

import (
	"fmt"
	"strings"

	"github.com/joshmedeski/sesh/v2/model"
)

// activeWindowNameFormat returns only the active window's name for each tmux
// session. Inactive windows render as an empty string and are ignored by
// ListAllWindowNames' parser.
const activeWindowNameFormat = "#{?window_active,#{window_name},}"

var listColorCodes = map[string]int{
	"black":          30,
	"red":            31,
	"green":          32,
	"yellow":         33,
	"blue":           34,
	"magenta":        35,
	"cyan":           36,
	"white":          37,
	"bright-black":   90,
	"gray":           90,
	"grey":           90,
	"bright-red":     91,
	"bright-green":   92,
	"bright-yellow":  93,
	"bright-blue":    94,
	"bright-magenta": 95,
	"bright-cyan":    96,
	"bright-white":   97,
}

func listColorCode(color string) (int, bool) {
	if color == "" {
		return 0, true
	}
	code, ok := listColorCodes[strings.ToLower(strings.TrimSpace(color))]
	return code, ok
}

func renderListFormatColors(format string, noColor bool) (string, error) {
	const prefix = "{fg:"

	var out strings.Builder
	for {
		start := strings.Index(format, prefix)
		if start == -1 {
			out.WriteString(format)
			break
		}

		out.WriteString(format[:start])
		format = format[start:]

		end := strings.IndexByte(format, '}')
		if end == -1 {
			return "", fmt.Errorf("unterminated foreground color token in --format")
		}

		token := format[:end+1]
		color := strings.TrimSpace(token[len(prefix) : len(token)-1])
		code, ok := listColorCode(color)
		if !ok || code == 0 {
			return "", fmt.Errorf("unsupported format color %q (use black, red, green, yellow, blue, magenta, cyan, white, or a bright-* variant)", color)
		}

		if !noColor {
			fmt.Fprintf(&out, "\033[%dm", code)
		}
		format = format[end+1:]
	}

	result := out.String()
	if noColor {
		return strings.ReplaceAll(result, "{/fg}", ""), nil
	}
	return strings.ReplaceAll(result, "{/fg}", "\033[39m"), nil
}

func sourceIconExcluded(excluded []string, source string) bool {
	for _, candidate := range excluded {
		if strings.EqualFold(strings.TrimSpace(candidate), source) {
			return true
		}
	}
	return false
}

func listFormatUsesActiveWindowName(format string) bool {
	return strings.Contains(format, "{active_window_name}") || strings.Contains(format, "{active_window_name_prefix}")
}

func hasTmuxSessions(sessions model.SeshSessions) bool {
	for _, key := range sessions.OrderedIndex {
		if sessions.Directory[key].Src == "tmux" {
			return true
		}
	}
	return false
}

func firstActiveWindowNameBySession(windowNames map[string][]string) map[string]string {
	if len(windowNames) == 0 {
		return nil
	}

	active := make(map[string]string, len(windowNames))
	for session, names := range windowNames {
		if len(names) > 0 {
			active[session] = names[0]
		}
	}
	return active
}

func formatListSession(format string, session model.SeshSession, displayName, activeWindowName string) string {
	activeWindowNamePrefix := ""
	if activeWindowName != "" {
		activeWindowNamePrefix = activeWindowName + " "
	}

	return strings.NewReplacer(
		"{name}", displayName,
		"{session}", session.Name,
		"{source}", session.Src,
		"{path}", session.Path,
		"{active_window_name}", activeWindowName,
		"{active_window_name_prefix}", activeWindowNamePrefix,
	).Replace(format)
}
