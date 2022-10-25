package args

import "fmt"

func Read(args []string) (string, string, error) {
	// Check if enough input parameters
	if len(args) < 3 {
		return "", "", fmt.Errorf("not enough input parameters")
	}

	token := args[1]
	currentVersion := args[2]
	return token, currentVersion, nil
}
