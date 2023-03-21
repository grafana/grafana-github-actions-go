package changelog

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFileUpdater(t *testing.T) {
	body := ChangelogBody{}
	body.Version = "9.4.4"
	tests := []struct {
		name           string
		input          string
		expectedOutput string
	}{
		{
			name:           "empty-file",
			input:          "",
			expectedOutput: "<!-- 9.4.4 START -->\n\n# 9.4.4\n\n<!-- 9.4.4 END -->\n",
		},
		{
			name:           "in-between",
			input:          "<!-- 9.5.0 START -->\n\n# 9.5.0\n\n<!-- 9.5.0 END -->\n<!-- 9.4.0 START -->\n\n# 9.4.0\n\n<!-- 9.4.0 END -->\n",
			expectedOutput: "<!-- 9.5.0 START -->\n\n# 9.5.0\n\n<!-- 9.5.0 END -->\n<!-- 9.4.4 START -->\n\n# 9.4.4\n\n<!-- 9.4.4 END -->\n<!-- 9.4.0 START -->\n\n# 9.4.0\n\n<!-- 9.4.0 END -->\n",
		},
		{
			name:           "newest",
			input:          "<!-- 9.4.1 START -->\n\n# 9.4.1\n\n<!-- 9.4.1 END -->\n<!-- 9.4.0 START -->\n\n# 9.4.0\n\n<!-- 9.4.0 END -->\n",
			expectedOutput: "<!-- 9.4.4 START -->\n\n# 9.4.4\n\n<!-- 9.4.4 END -->\n<!-- 9.4.1 START -->\n\n# 9.4.1\n\n<!-- 9.4.1 END -->\n<!-- 9.4.0 START -->\n\n# 9.4.0\n\n<!-- 9.4.0 END -->\n",
		},
		{
			name:           "replacing",
			input:          "<!-- 9.4.4 START -->\n\n# something\n\n<!-- 9.4.4 END -->\n<!-- 9.4.1 START -->\n\n# 9.4.1\n\n<!-- 9.4.1 END -->\n<!-- 9.4.0 START -->\n\n# 9.4.0\n\n<!-- 9.4.0 END -->\n",
			expectedOutput: "<!-- 9.4.4 START -->\n\n# 9.4.4\n\n<!-- 9.4.4 END -->\n<!-- 9.4.1 START -->\n\n# 9.4.1\n\n<!-- 9.4.1 END -->\n<!-- 9.4.0 START -->\n\n# 9.4.0\n\n<!-- 9.4.0 END -->\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			in := bytes.NewBufferString(test.input)
			out := bytes.Buffer{}
			require.NoError(t, UpdateFile(context.Background(), &out, in, &body))
			require.Equal(t, test.expectedOutput, out.String())
		})
	}
}
