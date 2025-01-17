package versions

import (
	"errors"
	"fmt"
	"strings"
)

func BumpReleaseBranch(branch string) (string, error) {
	if !strings.HasPrefix(branch, "release-") {
		return "", errors.New("release branches must have release- prefix")
	}

	v, err := Parse(strings.TrimPrefix(branch, "release-"))
	if err != nil {
		return "", err
	}

	v, err = BumpPatch(v)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("release-%s", v.String()), nil
}
