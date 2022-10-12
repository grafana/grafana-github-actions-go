package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadArg(t *testing.T) {
	// If there are less than 3 args, return an err
	t.Run("If there are less than 3 arguments, return an error", func(t *testing.T) {
		token, currentVersion, err := ReadArgs([]string{})
		require.Empty(t, token)
		require.Empty(t, currentVersion)
		require.NotNil(t, err)
	})
	// If there are correct amount of args, return them
	t.Run("If there are 3 or more arguments, return them", func(t *testing.T) {
		token, currentVersion, err := ReadArgs([]string{"/bin/go", "1234", "version"})
		require.Equal(t, "1234", token)
		require.Equal(t, "version", currentVersion)
		require.Nil(t, err)
	})
}
