package seshcli

import (
	"testing"

	"github.com/joshmedeski/sesh/v2/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActiveWindowNameFormat(t *testing.T) {
	assert.Equal(t, "#{?window_active,#{window_name},}", activeWindowNameFormat)
}

func TestListColorCode(t *testing.T) {
	tests := []struct {
		name  string
		color string
		code  int
		ok    bool
	}{
		{"standard color", "blue", 34, true},
		{"bright color", "bright-blue", 94, true},
		{"case and whitespace", "  Bright-Magenta  ", 95, true},
		{"gray alias", "gray", 90, true},
		{"grey alias", "grey", 90, true},
		{"empty color", "", 0, true},
		{"unsupported color", "orange", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, ok := listColorCode(tt.color)
			assert.Equal(t, tt.code, code)
			assert.Equal(t, tt.ok, ok)
		})
	}
}

func TestRenderListFormatColors(t *testing.T) {
	t.Run("renders foreground color and reset", func(t *testing.T) {
		got, err := renderListFormatColors("{fg:blue}hello{/fg}", false)
		require.NoError(t, err)
		assert.Equal(t, "\x1b[34mhello\x1b[39m", got)
	})

	t.Run("renders multiple scopes", func(t *testing.T) {
		got, err := renderListFormatColors(
			"{fg:red}red{/fg} {fg:bright-blue}blue{/fg}",
			false,
		)
		require.NoError(t, err)
		assert.Equal(t, "\x1b[31mred\x1b[39m \x1b[94mblue\x1b[39m", got)
	})

	t.Run("removes color tokens with no color", func(t *testing.T) {
		got, err := renderListFormatColors("{fg:blue}hello{/fg}", true)
		require.NoError(t, err)
		assert.Equal(t, "hello", got)
	})

	t.Run("preserves list placeholders", func(t *testing.T) {
		got, err := renderListFormatColors(
			"{fg:blue}{active_window_name_prefix}{/fg}{name}",
			false,
		)
		require.NoError(t, err)
		assert.Equal(
			t,
			"\x1b[34m{active_window_name_prefix}\x1b[39m{name}",
			got,
		)
	})

	t.Run("rejects unsupported colors", func(t *testing.T) {
		_, err := renderListFormatColors("{fg:orange}hello{/fg}", false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), `unsupported format color "orange"`)
	})

	t.Run("rejects empty colors", func(t *testing.T) {
		_, err := renderListFormatColors("{fg:}hello{/fg}", false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), `unsupported format color ""`)
	})

	t.Run("rejects unterminated tokens", func(t *testing.T) {
		_, err := renderListFormatColors("{fg:blue", false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unterminated foreground color token")
	})
}

func TestSourceIconExcluded(t *testing.T) {
	excluded := []string{"tmux", " config "}

	assert.True(t, sourceIconExcluded(excluded, "tmux"))
	assert.True(t, sourceIconExcluded(excluded, "TMUX"))
	assert.True(t, sourceIconExcluded(excluded, "config"))
	assert.False(t, sourceIconExcluded(excluded, "zoxide"))
	assert.False(t, sourceIconExcluded(nil, "tmux"))
}

func TestListFormatUsesActiveWindowName(t *testing.T) {
	assert.True(t, listFormatUsesActiveWindowName("{active_window_name} {name}"))
	assert.True(t, listFormatUsesActiveWindowName("{active_window_name_prefix}{name}"))
	assert.False(t, listFormatUsesActiveWindowName("{source} {name}"))
}

func TestHasTmuxSessions(t *testing.T) {
	t.Run("true when tmux session is present", func(t *testing.T) {
		sessions := model.SeshSessions{
			OrderedIndex: []string{"config:notes", "tmux:work"},
			Directory: model.SeshSessionMap{
				"config:notes": {Src: "config", Name: "notes"},
				"tmux:work":    {Src: "tmux", Name: "work"},
			},
		}
		assert.True(t, hasTmuxSessions(sessions))
	})

	t.Run("false when no tmux session is present", func(t *testing.T) {
		sessions := model.SeshSessions{
			OrderedIndex: []string{"config:notes", "zoxide:work"},
			Directory: model.SeshSessionMap{
				"config:notes": {Src: "config", Name: "notes"},
				"zoxide:work":  {Src: "zoxide", Name: "work"},
			},
		}
		assert.False(t, hasTmuxSessions(sessions))
	})
}

func TestFirstActiveWindowNameBySession(t *testing.T) {
	t.Run("uses first returned name", func(t *testing.T) {
		got := firstActiveWindowNameBySession(map[string][]string{
			"work":     {"nvim", "shell"},
			"dotfiles": {"zsh"},
			"empty":    {},
		})
		assert.Equal(t, map[string]string{
			"work":     "nvim",
			"dotfiles": "zsh",
		}, got)
	})

	t.Run("nil for no window names", func(t *testing.T) {
		assert.Nil(t, firstActiveWindowNameBySession(nil))
		assert.Nil(t, firstActiveWindowNameBySession(map[string][]string{}))
	})
}

func TestFormatListSession(t *testing.T) {
	session := model.SeshSession{
		Src:  "tmux",
		Name: "work",
		Path: "/home/user/work",
	}

	t.Run("replaces all supported fields", func(t *testing.T) {
		got := formatListSession(
			"{source}|{session}|{path}|{active_window_name}|{active_window_name_prefix}{name}",
			session,
			" work",
			"",
		)
		assert.Equal(t, "tmux|work|/home/user/work||  work", got)
	})

	t.Run("prefix has one trailing space", func(t *testing.T) {
		got := formatListSession(
			"{active_window_name_prefix}{name}",
			session,
			"work",
			"",
		)
		assert.Equal(t, " work", got)
	})

	t.Run("empty prefix disappears completely", func(t *testing.T) {
		got := formatListSession(
			"{active_window_name_prefix}{name}",
			session,
			"work",
			"",
		)
		assert.Equal(t, "work", got)
	})

	t.Run("active window name has no implicit spacing", func(t *testing.T) {
		got := formatListSession(
			"{active_window_name}:{name}",
			session,
			"work",
			"",
		)
		assert.Equal(t, ":work", got)
	})

	t.Run("works with inline colors", func(t *testing.T) {
		format, err := renderListFormatColors(
			"{fg:blue}{active_window_name_prefix}{/fg}{name}",
			false,
		)
		require.NoError(t, err)

		got := formatListSession(format, session, "work", "")
		assert.Equal(t, "\x1b[34m \x1b[39mwork", got)
	})
}
