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
		content := bytes.NewBufferString(`### Bug fixes

- **Category:** some title. [#123](https://github.com/grafana/grafana/issue/123), [@user](https://github.com/user)
- **Other Category:** some other title. [#124](https://github.com/grafana/grafana/issue/124), [@user](https://github.com/user)
- **Enterprise:** some other title. (Enterprise)
- **Some HTML:** some <summary>. (Enterprise)
`)
		expectedEntries := []string{
			"Category: some title.",
			"Other Category: some other title.",
			"Enterprise: some other title. (Enterprise)",
			"Some HTML: some <summary>. (Enterprise)",
		}
		result, err := p.Parse(ctx, content)
		require.NoError(t, err)
		require.Len(t, result, 1)
		require.Equal(t, "Bug fixes", result[0].Title)
		require.Len(t, result[0].Entries, len(expectedEntries))
		entries := result[0].Entries
		for idx, expected := range expectedEntries {
			require.Equal(t, expected, entries[idx].Title)
		}
	})
}
