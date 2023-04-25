package community

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/rs/zerolog"
)

type Community struct {
	key        string
	username   string
	baseURL    string
	httpClient http.Client
}

func New(options ...CommunityOption) *Community {
	c := &Community{
		httpClient: http.Client{},
	}
	for _, opt := range options {
		opt(c)
	}
	return c
}

type PostInput struct {
	Title    string `json:"title"`
	Body     string `json:"raw"`
	Category int    `json:"category"`
	Author   string `json:"-"`
}

// CreateOrUpdate tries to update an existing post with the provided title in
// the specified category. If no such topic exists, a new topic with the same
// content will be created.
func (c *Community) CreateOrUpdatePost(ctx context.Context, post PostInput) (int, error) {
	logger := zerolog.Ctx(ctx)
	category, err := c.getCategory(ctx, post.Category)
	if err != nil {
		return -1, err
	}
	searchQuery := fmt.Sprintf("%s @%s #%s in:title order:latest_topic", post.Title, post.Author, category.Slug)
	opts := SearchOptions{
		Page: 1,
	}
	result, err := c.search(ctx, searchQuery, &opts)
	if err != nil {
		return -1, err
	}
	if len(result.Posts) > 0 {
		// No post found, so let's create a new one
		logger.Info().Msgf("Updating post %d", result.Posts[0].ID)
		return result.Posts[0].ID, c.updatePost(ctx, result.Posts[0].ID, post.Body)
	}

	topic, err := c.createTopic(ctx, post)
	if err != nil {
		return -1, err
	}
	return topic.ID, nil
}

func (c *Community) createTopic(ctx context.Context, post PostInput) (*Post, error) {
	body := bytes.Buffer{}
	result := Post{}
	if err := json.NewEncoder(&body).Encode(post); err != nil {
		return nil, err
	}
	req, err := c.buildRequest(ctx, http.MethodPost, "/posts.json", nil, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		io.Copy(os.Stderr, resp.Body)
		return nil, fmt.Errorf("creating a new post failed")
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Community) updatePost(ctx context.Context, postID int, raw string) error {
	body := bytes.Buffer{}
	input := map[string]any{
		"post": map[string]any{
			"raw":         raw,
			"edit_reason": "Changelog was updated",
		},
	}
	if err := json.NewEncoder(&body).Encode(input); err != nil {
		return err
	}
	req, err := c.buildRequest(ctx, http.MethodPut, fmt.Sprintf("/posts/%d.json", postID), nil, &body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		io.Copy(os.Stderr, resp.Body)
		return fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}
	return nil
}

func (c *Community) getCategory(ctx context.Context, id int) (*Category, error) {
	result := Category{}
	req, err := c.buildRequest(ctx, http.MethodGet, fmt.Sprintf("/c/%d/show.json", id), nil, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Community) buildRequest(ctx context.Context, method string, path string, values url.Values, body io.Reader) (*http.Request, error) {
	fullPath := strings.Builder{}
	fullPath.WriteString(c.baseURL)
	fullPath.WriteString(path)
	if len(values) > 0 {
		fullPath.WriteString("?")
		fullPath.WriteString(values.Encode())
	}
	req, err := http.NewRequestWithContext(ctx, method, fullPath.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Api-Username", c.username)
	req.Header.Set("Api-Key", c.key)
	return req, nil
}
