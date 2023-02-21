package node

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type PackageJSON struct {
	Bin         string `json:"bin"`
	Description string `json:"description"`
}

func ParsePackageJson(p string) (*PackageJSON, error) {
	bytes, err := os.ReadFile(p)

	if err != nil {
		return nil, err
	}

	var pkg PackageJSON
	err = json.Unmarshal(bytes, &pkg)
	return &pkg, err
}

func FindPackageLock(root string) (*string, error) {
	errorsSeen := []string{}

	candidates := []string{
		"package-lock.json",
		"yarn.lock",
		"pnpm-lock.yaml",
	}

	for _, candidate := range candidates {
		candidatePath := filepath.Join(root, candidate)
		_, err := os.Open(candidatePath)
		if err != nil {
			errorsSeen = append(errorsSeen, err.Error())
		} else {
			return &candidatePath, nil
		}
	}

	return nil, fmt.Errorf("no lockfile found underneath `%s`. \n%s", root, strings.Join(errorsSeen, "\n"))
}
