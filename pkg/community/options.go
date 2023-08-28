package community

import "net/http"

type CommunityOption func(c *Community)

func CommunityWithBaseURL(u string) CommunityOption {
	return func(c *Community) {
		c.baseURL = u
	}
}

func CommunityWithAPICredentials(username, key string) CommunityOption {
	return func(c *Community) {
		c.key = key
		c.username = username
	}
}

func CommunityWithHTTPClient(client *http.Client) CommunityOption {
	return func(c *Community) {
		c.httpClient = client
	}
}
