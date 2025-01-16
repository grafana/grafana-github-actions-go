package versions

import (
	"errors"
	"fmt"
	"strings"
)

func VersionBranch(version string) (string, error) {
	version = strings.TrimPrefix(version, "v")
	groups := semverRegexp.FindStringSubmatch(version)
	// The first group is the entire string, so we need 3 results
	if len(groups) < 3 {
		return "", errors.New("version does not match a semver regex")
	}

	return fmt.Sprintf("v%s.%s.x", groups[1], groups[2]), nil
}
