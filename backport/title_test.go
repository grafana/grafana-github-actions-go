package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatBackportTitle(t *testing.T) {
	t.Run("default template when empty", func(t *testing.T) {
		assert.Equal(t, "[release-12.0.0] Example Bug Fix", FormatBackportTitle("", "release-12.0.0", "Example Bug Fix"))
	})

	t.Run("explicit default template", func(t *testing.T) {
		assert.Equal(t, "[gel-release-3.6] fix: something", FormatBackportTitle("[{{branch}}] {{title}}", "gel-release-3.6", "fix: something"))
	})

	t.Run("suffix template", func(t *testing.T) {
		assert.Equal(t, "fix: something [gel-release-3.6]", FormatBackportTitle("{{title}} [{{branch}}]", "gel-release-3.6", "fix: something"))
	})
}
