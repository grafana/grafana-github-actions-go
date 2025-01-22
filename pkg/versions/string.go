package versions

import (
	"errors"
	"fmt"
	"strconv"
)

type Version struct {
	Major      string
	Minor      string
	Patch      string
	Prerelease string
	Buildmeta  string
}

func (v Version) String() string {
	base := fmt.Sprintf("%s.%s.%s", v.Major, v.Minor, v.Patch)
	if v.Prerelease != "" {
		base += fmt.Sprintf("-%s", v.Prerelease)
	}

	if v.Buildmeta != "" {
		base += fmt.Sprintf("+%s", v.Buildmeta)
	}

	return base
}

func Parse(v string) (Version, error) {
	matches := SemverRegexp.FindStringSubmatch(v)
	if len(matches) < 3 {
		return Version{}, errors.New("version does not match a semver regex")
	}
	groups := make(map[string]string)
	for i, name := range SemverRegexp.SubexpNames() {
		if i > 0 && i <= len(matches) {
			groups[name] = matches[i]
		}
	}

	return Version{
		Major:      groups["major"],
		Minor:      groups["minor"],
		Patch:      groups["patch"],
		Prerelease: groups["prerelease"],
		Buildmeta:  groups["buildmetadata"],
	}, nil
}

// BumpMinor bumps the minor version, resets the patch version to 0, and removes prerelease and buildmeta.
func BumpMinor(v Version) (Version, error) {
	minor, err := strconv.ParseInt(v.Minor, 10, 64)
	if err != nil {
		return Version{}, err
	}

	minor = minor + 1

	return Version{
		Major: v.Major,
		Minor: strconv.FormatInt(minor, 10),
		Patch: "0",
	}, nil
}

// BumpPatch bumps the patch version and removes prerelease and buildmeta.
func BumpPatch(v Version) (Version, error) {
	patch, err := strconv.ParseInt(v.Patch, 10, 64)
	if err != nil {
		return Version{}, err
	}

	patch = patch + 1

	return Version{
		Major: v.Major,
		Minor: v.Minor,
		Patch: strconv.FormatInt(patch, 10),
	}, nil
}
