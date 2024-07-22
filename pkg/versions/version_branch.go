package versions

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

func VersionBranch(version string) (string, error) {
	version = strings.TrimPrefix(version, "v")
	r, err := regexp.Compile(semverRegex)
	if err != nil {
		return "", err
	}

	groups := r.FindStringSubmatch(version)
	// The first group is the entire string, so we need 3 results
	if len(groups) < 3 {
		return "", errors.New("version does not match a semver regex")
	}

	return fmt.Sprintf("v%s.%s.x", groups[1], groups[2]), nil
}
