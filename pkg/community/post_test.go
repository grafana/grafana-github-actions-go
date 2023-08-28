package community

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
)

func TestCommunityPost(t *testing.T) {
	t.Run("too-much-content-create", func(t *testing.T) {
		// If the changelog is larger than 50000 characters, then the server
		// will respond with an error code 422 and the following message:
		//
		// `{"action":"create_post","errors":["Body is limited to 50000 characters; you entered [...]."]}`
		//
		// In such a situation, the client should try again with a shorter
		// message.
		ctx := context.Background()
		srv := NewMockCommunityServer()
		comm := New(CommunityWithBaseURL(srv.GetURL()), CommunityWithHTTPClient(srv.GetClient()))

		// This produces a string with 50005 characters, which will go beyond
		// the size limit:
		body := strings.Repeat("hello", 10001)
		_, err := comm.CreateOrUpdatePost(ctx, PostInput{
			Title:    "Sample Post",
			Category: 4,
			Body:     body,
			Author:   "test",
		}, &PostOptions{
			FallbackBody: "fallback",
		})
		require.NoError(t, err)
		postCalls := srv.GetPostCalls()
		require.Len(t, postCalls, 2)
		require.Equal(t, "fallback", postCalls[1].Raw)
	})

	t.Run("too-much-content-update", func(t *testing.T) {
		// Same as the previous test but for updating an existing post.
		ctx := context.Background()
		srv := NewMockCommunityServer()
		srv.ExistingPosts = []TestSearchPost{
			{
				ID:      1,
				TopicID: 1,
			},
		}
		comm := New(CommunityWithBaseURL(srv.GetURL()), CommunityWithHTTPClient(srv.GetClient()))
		body := strings.Repeat("hello", 10001)
		_, err := comm.CreateOrUpdatePost(ctx, PostInput{
			Title:    "Sample Post",
			Category: 4,
			Body:     body,
			Author:   "test",
		}, &PostOptions{
			FallbackBody: "fallback",
		})
		require.NoError(t, err)
		postCalls := srv.GetPostCalls()
		require.Len(t, postCalls, 2)
		require.Equal(t, "fallback", postCalls[1].Raw)
	})
}

type TestPost struct {
	Title    string
	Raw      string
	Category int
	Author   string
}
type TestPostUpdate struct {
	Post struct {
		Raw string
	}
}

type TestSearchPost struct {
	ID      int `json:"id"`
	TopicID int `json:"topic_id"`
}

type MockCommunityServer struct {
	srv           *httptest.Server
	PostSizeLimit int
	ExistingPosts []TestSearchPost
	postCalls     []TestPost
}

func NewMockCommunityServer() *MockCommunityServer {
	s := MockCommunityServer{
		PostSizeLimit: 50000,
		postCalls:     make([]TestPost, 0, 5),
	}
	handler := http.NewServeMux()
	handler.HandleFunc("/c/4/show.json", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"category": {"id": 4, "slug": "hello"}}`)
	})
	handler.HandleFunc("/search.json", func(w http.ResponseWriter, r *http.Request) {
		data := map[string]interface{}{
			"posts": s.ExistingPosts,
		}
		json.NewEncoder(w).Encode(data)
	})
	handler.HandleFunc("/posts.json", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		post := TestPost{}
		if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
			http.Error(w, "Decoding failed", http.StatusInternalServerError)
			return
		}
		s.postCalls = append(s.postCalls, post)
		if size := utf8.RuneCountInString(post.Raw); size > s.PostSizeLimit {
			w.WriteHeader(http.StatusUnprocessableEntity)
			fmt.Fprintf(w, `{"action":"create_post","errors":["Body is limited to %d characters; you entered %d."]}`, s.PostSizeLimit, size)
			return
		}
		fmt.Fprintf(w, `{}`)
	})
	handler.HandleFunc("/posts/1.json", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		post := TestPostUpdate{}
		if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
			http.Error(w, "Decoding failed", http.StatusInternalServerError)
			return
		}
		s.postCalls = append(s.postCalls, TestPost{
			Raw: post.Post.Raw,
		})
		if size := utf8.RuneCountInString(post.Post.Raw); size > s.PostSizeLimit {
			w.WriteHeader(http.StatusUnprocessableEntity)
			fmt.Fprintf(w, `{"action":"create_post","errors":["Body is limited to %d characters; you entered %d."]}`, s.PostSizeLimit, size)
			return
		}
		fmt.Fprintf(w, `{}`)
	})
	srv := httptest.NewServer(handler)
	s.srv = srv
	return &s
}

func (m *MockCommunityServer) GetPostCalls() []TestPost {
	return m.postCalls
}

func (m *MockCommunityServer) GetURL() string {
	return m.srv.URL
}

func (m *MockCommunityServer) GetClient() *http.Client {
	return m.srv.Client()
}
