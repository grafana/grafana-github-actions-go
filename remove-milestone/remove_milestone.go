package main

import (
	"context"
	"fmt"
	"os"

	gh "github.com/google/go-github/v47/github"
	"golang.org/x/oauth2"
)

func main() {
	// we need to get all open issue with milestone and remove the milestone from them
	// we need to get all PR opened with milestone and remove the milestone from them
	fmt.Println("Make it workkkkk")
	if len(os.Args) <= 1 {
		fmt.Println("Not enough input parameters")
		os.Exit(1)
	}
	// Just using something simple to dmeonstrate using the github package here
	argsWithoutProg := os.Args[1:]

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: argsWithoutProg[0]},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := gh.NewClient(tc)

	// list all repositories for the authenticated user
	_, _, err := client.Repositories.List(ctx, "", nil)
	if err != nil {
		fmt.Println("the list of repostories failed")
		os.Exit(1)
	}
}
