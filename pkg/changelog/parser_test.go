package changelog

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	ctx := context.Background()

	t.Run("empty file", func(t *testing.T) {
		p := NewParser()
		content := bytes.NewBufferString("")
		result, err := p.Parse(ctx, content)
		require.NoError(t, err)
		require.Len(t, result, 0)
	})

	t.Run("simple file", func(t *testing.T) {
		p := NewParser()
		content := bytes.NewBufferString(`## Bug fixes

- **Category**: some title. [#123](https://github.com/grafana/grafana/issue/123), [@user](https://github.com/user)
- **Other Category**: some other title. [#124](https://github.com/grafana/grafana/issue/124), [@user](https://github.com/user)
`)
		result, err := p.Parse(ctx, content)
		require.NoError(t, err)
		require.Len(t, result, 1)
		require.Equal(t, "Bug fixes", result[0].Title)
		require.Len(t, result[0].Entries, 2)
		entries := result[0].Entries
		require.Equal(t, "Category: some title.", entries[0].Title)
		require.Equal(t, "Other Category: some other title.", entries[1].Title)
	})
}
