package toolkit

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestSubmitMetrics(t *testing.T) {
	ctx := context.Background()
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr})
	ctx = logger.WithContext(ctx)

	t.Setenv("GITHUB_TOKEN", "12345")

	t.Run("no-submission-without-creds", func(t *testing.T) {
		submitted := false
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			submitted = true
		}))
		t.Cleanup(srv.Close)
		t.Setenv("INPUT_METRICS_API_ENDPOINT", srv.URL)
		t.Setenv("INPUT_METRICS_API_USERNAME", "")
		t.Setenv("INPUT_METRICS_API_KEY", "")
		t.Setenv("GITHUB_REPOSITORY", "grafana/gha-testing")
		tk, err := Init(ctx)
		require.NoError(t, err)
		err = tk.submitMetrics(ctx, []Metric{
			{
				Name:  "dummy",
				Value: 123.0,
			},
		})
		require.NoError(t, err)
		require.False(t, submitted)
	})

	t.Run("submission-with-creds", func(t *testing.T) {
		received := make([]graphiteMetric, 0, 5)
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			m := []graphiteMetric{}
			defer r.Body.Close()
			require.NoError(t, json.NewDecoder(r.Body).Decode(&m))
			received = append(received, m...)
		}))
		t.Cleanup(srv.Close)
		t.Setenv("INPUT_METRICS_API_ENDPOINT", srv.URL)
		t.Setenv("INPUT_METRICS_API_USERNAME", "username")
		t.Setenv("INPUT_METRICS_API_KEY", "password")
		t.Setenv("GITHUB_REPOSITORY", "grafana/gha-testing")
		tk, err := Init(ctx)
		require.NoError(t, err)
		err = tk.submitMetrics(ctx, []Metric{
			{
				Name:  "dummy",
				Value: 123.0,
			},
		})
		require.NoError(t, err)
		require.Len(t, received, 1)

		// Verify that the metrics are prefixed correctly:
		for _, m := range received {
			require.Equal(t, "repo_stats.gha-testing.dummy", m.Name)
		}
	})

	t.Run("submission-with-creds-grafana", func(t *testing.T) {
		received := make([]graphiteMetric, 0, 5)
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			m := []graphiteMetric{}
			defer r.Body.Close()
			require.NoError(t, json.NewDecoder(r.Body).Decode(&m))
			received = append(received, m...)
		}))
		t.Cleanup(srv.Close)
		t.Setenv("INPUT_METRICS_API_ENDPOINT", srv.URL)
		t.Setenv("INPUT_METRICS_API_USERNAME", "username")
		t.Setenv("INPUT_METRICS_API_KEY", "password")
		t.Setenv("GITHUB_REPOSITORY", "grafana/grafana")
		tk, err := Init(ctx)
		require.NoError(t, err)
		err = tk.submitMetrics(ctx, []Metric{
			{
				Name:  "dummy",
				Value: 123.0,
			},
		})
		require.NoError(t, err)
		require.Len(t, received, 1)

		// Verify that the metrics are prefixed correctly:
		for _, m := range received {
			require.Equal(t, "gh_action.dummy", m.Name)
		}
	})
}
