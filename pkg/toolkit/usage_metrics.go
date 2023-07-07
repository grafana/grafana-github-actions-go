package toolkit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/go-github/v50/github"
	"github.com/rs/zerolog"
)

type Metric struct {
	Name  string
	Value float64
}

func (m Metric) String() string {
	return fmt.Sprintf("%s=%f", m.Name, m.Value)
}

func (tk *Toolkit) submitMetrics(ctx context.Context, metrics []Metric) error {
	logger := zerolog.Ctx(ctx)
	if tk.metricsAPIKey == "" || tk.metricsAPIUsername == "" {
		logger.Info().Msg("Metric submission disabled (no API username/key set)")
		return nil
	}
	for _, metric := range metrics {
		if err := tk.trackMetric(ctx, metric); err != nil {
			return fmt.Errorf("failed to submit metric %s: %w", metric.Name, err)
		}
	}
	return nil
}

// SubmitUsageMetrics sends metrics exposed by the GitHub rate limiter et al. to
// a configured Graphite HTTP endpoint. If username or api key for that endpoint
// are empty, this operation is a no-op.
func (tk *Toolkit) SubmitUsageMetrics(ctx context.Context) error {
	limits, _, err := tk.ghClient.RateLimits(ctx)
	if err != nil {
		return fmt.Errorf("failed to retrieve rate limits from GitHub: %w", err)
	}

	metrics := []Metric{
		{
			Name:  "octokit_request_count",
			Value: float64(tk.requestCount.Load()),
		},
		{
			Name:  "usage_core",
			Value: calculateUsage(limits.Core),
		},
		{
			Name:  "usage_search",
			Value: calculateUsage(limits.Search),
		},
		{
			Name:  "usage_graphql",
			Value: calculateUsage(limits.GraphQL),
		},
	}
	return tk.submitMetrics(ctx, metrics)
}

func calculateUsage(rate *github.Rate) float64 {
	return 1.0 - float64(rate.Remaining)/float64(rate.Limit)
}

type graphiteMetric struct {
	Name     string   `json:"name"`
	Value    float64  `json:"value"`
	Interval int64    `json:"interval"`
	MType    string   `json:"mtype"`
	Time     int64    `json:"time"`
	Tags     []string `json:"tags"`
}

func (tk *Toolkit) trackMetric(ctx context.Context, metric Metric) error {
	now := time.Now()
	metricName := fmt.Sprintf("%s.%s", tk.metricsNamePrefix, metric.Name)
	metricName = strings.ReplaceAll(metricName, "/", "_")
	gm := []graphiteMetric{
		{
			Name:     metricName,
			Value:    metric.Value,
			Interval: 60,
			MType:    "count",
			Time:     now.Unix(),
			Tags:     []string{},
		},
	}
	httpClient := http.Client{}
	body := bytes.Buffer{}
	if err := json.NewEncoder(&body).Encode(gm); err != nil {
		return fmt.Errorf("failed to encode metrics: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tk.metricsAPIEndpoint, &body)
	if err != nil {
		return fmt.Errorf("failed to construct HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(tk.metricsAPIUsername, tk.metricsAPIKey)
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP requets: %w", err)
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("metrics request received unexpected status code: %d", resp.StatusCode)
	}
	return nil
}
