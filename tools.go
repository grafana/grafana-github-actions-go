//go:build tools
// +build tools

package client

import (
	_ "github.com/Khan/genqlient"
)

// NOTE: This file exists so that genqlient stays available across `go mod tidy` runs
