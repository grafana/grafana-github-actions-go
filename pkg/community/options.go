package community

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
