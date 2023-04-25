package community

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

type SearchResult struct {
	Posts []Post `json:"posts"`
}

type Post struct {
	ID      int `json:"id"`
	TopicID int `json:"topic_id"`
}

type Category struct {
	ID   int    `json:"id"`
	Slug string `json:"slug"`
}

type SearchOptions struct {
	Page int
}

func (c *Community) search(ctx context.Context, query string, opts *SearchOptions) (*SearchResult, error) {
	result := SearchResult{}
	qs := url.Values{}
	qs.Set("page", strconv.Itoa(opts.Page))
	qs.Set("q", query)
	req, err := c.buildRequest(ctx, http.MethodGet, "/search.json", qs, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request URL: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send search request: %w", err)
	}
	if resp.StatusCode != 200 {
		io.Copy(os.Stderr, resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("search request returned with unexpected status code %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}
	return &result, nil
}
