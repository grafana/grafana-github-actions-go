package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVersionExtraction(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectErr    bool
		expectOutput string
	}{
		{
			name:         "valid",
			input:        `{"version": "10.1.0-pre"}`,
			expectErr:    false,
			expectOutput: "10.1.x",
		},
		{
			name:         "no-field",
			input:        `{}`,
			expectErr:    true,
			expectOutput: "",
		},
		{
			name:         "invalid-version",
			input:        `{"version": "hello"}`,
			expectErr:    true,
			expectOutput: "",
		},
		{
			name:         "not-json",
			input:        `hello`,
			expectErr:    true,
			expectOutput: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			v, err := versionFromPackage(test.input)
			if err != nil && !test.expectErr {
				t.Fatalf("unexpected error: %s", err.Error())
			}
			if err == nil && test.expectErr {
				t.Fatalf("missing expected error")
			}
			require.Equal(t, test.expectOutput, v)
		})
	}
}
