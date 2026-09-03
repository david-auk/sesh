package seshcli

import (
	"testing"

	"github.com/joshmedeski/sesh/v2/model"
	"github.com/stretchr/testify/assert"
)

func TestListFormatUsesActiveWindowName(t *testing.T) {
	assert.True(t, listFormatUsesActiveWindowName("{active_window_name_prefix}{name}"))
	assert.True(t, listFormatUsesActiveWindowName("{name} [{active_window_name}]"))
	assert.False(t, listFormatUsesActiveWindowName("{source}: {name}"))
}

func TestHasTmuxSessions(t *testing.T) {
	sessions := model.SeshSessions{
		OrderedIndex: []string{"config:notes", "tmux:sesh"},
		Directory: model.SeshSessionMap{
			"config:notes": {Src: "config", Name: "notes"},
			"tmux:sesh":    {Src: "tmux", Name: "sesh"},
		},
	}
	assert.True(t, hasTmuxSessions(sessions))

	sessions.OrderedIndex = []string{"config:notes"}
	assert.False(t, hasTmuxSessions(sessions))
}

func TestFirstActiveWindowNameBySession(t *testing.T) {
	windowNames := map[string][]string{
		"sesh":     {""},
		"dotfiles": {""},
		"empty":    {},
	}

	assert.Equal(t, map[string]string{
		"sesh":     "",
		"dotfiles": "",
	}, firstActiveWindowNameBySession(windowNames))
	assert.Nil(t, firstActiveWindowNameBySession(nil))
}

func TestFormatListSession(t *testing.T) {
	session := model.SeshSession{
		Src:  "tmux",
		Name: "sesh",
		Path: "/home/user/code/sesh",
	}

	t.Run("formats all placeholders", func(t *testing.T) {
		got := formatListSession(
			"{active_window_name_prefix}{name} | {session} | {source} | {path} | {active_window_name}",
			session,
			" sesh",
			"",
		)
		assert.Equal(t, "  sesh | sesh | tmux | /home/user/code/sesh | ", got)
	})

	t.Run("active window name prefix disappears when there is no active window", func(t *testing.T) {
		got := formatListSession("{active_window_name_prefix}{name}", session, " sesh", "")
		assert.Equal(t, " sesh", got)
	})

	t.Run("unknown placeholders are preserved", func(t *testing.T) {
		got := formatListSession("{name} {unknown}", session, "sesh", "")
		assert.Equal(t, "sesh {unknown}", got)
	})
}

func TestListColorCode(t *testing.T) {
	code, ok := listColorCode("yellow")
	assert.True(t, ok)
	assert.Equal(t, 33, code)

	code, ok = listColorCode("Bright-Cyan")
	assert.True(t, ok)
	assert.Equal(t, 96, code)

	code, ok = listColorCode("")
	assert.True(t, ok)
	assert.Equal(t, 0, code)

	_, ok = listColorCode("chartreuse")
	assert.False(t, ok)
}

func TestRenderListFormatColors(t *testing.T) {
	t.Run("renders scoped foreground colors", func(t *testing.T) {
		got, err := renderListFormatColors("{fg:yellow}{active_window_name_prefix}{/fg}{name}", false)
		assert.NoError(t, err)
		assert.Equal(t, "\x1b[33m{active_window_name_prefix}\x1b[39m{name}", got)
	})

	t.Run("supports multiple colors", func(t *testing.T) {
		got, err := renderListFormatColors("{fg:gray}{source}{/fg} {fg:bright-cyan}{active_window_name}{/fg}", false)
		assert.NoError(t, err)
		assert.Equal(t, "\x1b[90m{source}\x1b[39m \x1b[96m{active_window_name}\x1b[39m", got)
	})

	t.Run("no color strips formatting tokens", func(t *testing.T) {
		got, err := renderListFormatColors("{fg:yellow}{active_window_name_prefix}{/fg}{name}", true)
		assert.NoError(t, err)
		assert.Equal(t, "{active_window_name_prefix}{name}", got)
	})

	t.Run("rejects unsupported colors", func(t *testing.T) {
		_, err := renderListFormatColors("{fg:chartreuse}{name}{/fg}", false)
		assert.EqualError(t, err, `unsupported format color "chartreuse" (use black, red, green, yellow, blue, magenta, cyan, white, or a bright-* variant)`)
	})

	t.Run("rejects unterminated color tokens", func(t *testing.T) {
		_, err := renderListFormatColors("{fg:yellow", false)
		assert.EqualError(t, err, "unterminated foreground color token in --format")
	})
}

func TestSourceIconExcluded(t *testing.T) {
	excluded := []string{"tmux", " config "}
	assert.True(t, sourceIconExcluded(excluded, "tmux"))
	assert.True(t, sourceIconExcluded(excluded, "CONFIG"))
	assert.False(t, sourceIconExcluded(excluded, "zoxide"))
}
