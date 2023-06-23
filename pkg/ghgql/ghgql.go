//go:generate go run github.com/Khan/genqlient
package ghgql

import (
	"net/http"

	"github.com/Khan/genqlient/graphql"
)

type Client struct {
	gql graphql.Client
}

type doerWithToken struct {
	token string
}

func (d *doerWithToken) Do(req *http.Request) (*http.Response, error) {
	req.Header.Add("Authorization", "bearer "+d.token)
	return http.DefaultClient.Do(req)
}

func NewClient(token string) *Client {
	gqlc := graphql.NewClient("https://api.github.com/graphql", &doerWithToken{token: token})
	return &Client{
		gql: gqlc,
	}
}
