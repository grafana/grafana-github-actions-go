package versions_test

import (
	"testing"

	"github.com/grafana/grafana-github-actions-go/pkg/versions"
	"github.com/stretchr/testify/require"
)

func TestString(t *testing.T) {
	results := []string{
		"1.2.3+security-01",
		"1.2.3-1+security-01",
		"1.2.3",
		"1.2.3-1",
		"1.2.300-1",
		"100.200.300-pre1",
		"100.200.300-pre1+example-build-meta",
	}

	for _, v := range results {
		t.Run(v, func(t *testing.T) {
			r, err := versions.Parse(v)
			require.NoError(t, err)
			require.Equal(t, v, r.String())
		})
	}
}
