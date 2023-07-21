package changelog

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractContentForVersion(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		version        string
		opts           *ExtractContentOptions
		expectedOutput string
		expectedFound  bool
		expectedError  bool
	}{
		{
			name:           "empty input",
			version:        "v1.2.3",
			input:          "",
			expectedOutput: "",
			expectedFound:  false,
			expectedError:  false,
		},
		{
			name:           "content present without heading somewhere in the file",
			version:        "v1.2.3",
			input:          "<!-- 1.2.2 START -->\nwrong content\n<!-- 1.2.2 END -->\n<!-- 1.2.3 START -->\ncontent\n<!-- 1.2.3 END -->\n<!-- 1.2.4 START -->\nwrong content\n<!-- 1.2.4 END -->",
			expectedOutput: "content",
			expectedFound:  true,
			expectedError:  false,
		},
		{
			name:           "content present without heading",
			version:        "v1.2.3",
			input:          "<!-- 1.2.3 START -->\ncontent\n<!-- 1.2.3 END -->",
			expectedOutput: "content",
			expectedFound:  true,
			expectedError:  false,
		},
		{
			name:    "content present with heading removed",
			version: "v1.2.3",
			input:   "<!-- 1.2.3 START -->\n\n\n# 1.2.3 (2023-07-08)\n\ncontent\n<!-- 1.2.3 END -->",
			opts: &ExtractContentOptions{
				RemoveHeadling: true,
			},
			expectedOutput: "content",
			expectedFound:  true,
			expectedError:  false,
		},
		{
			name:           "content present with multiple lines",
			version:        "v1.2.3",
			input:          "<!-- 1.2.3 START -->\ncontent\nsecond line\n<!-- 1.2.3 END -->",
			expectedOutput: "content\nsecond line",
			expectedFound:  true,
			expectedError:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			output, found, err := ExtractContentForVersion(ctx, bytes.NewBufferString(test.input), test.version, test.opts)
			if test.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expectedFound, found)
				require.Equal(t, test.expectedOutput, output)
			}
		})
	}
}
