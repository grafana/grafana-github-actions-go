package main

import (
	"context"
	"strings"
	"testing"
)

type MockClient struct {
	EditPRCalled        bool
	CreateCommentCalled bool
	ShouldError         bool
	PrevBranch          string
	PrevComment         string
}
