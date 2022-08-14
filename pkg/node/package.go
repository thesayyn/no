package node

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type PackageJSON struct {
	Bin         string `json:"bin"`
	Description string `json:"description"`
}

func GetPackageJson(root string) (*PackageJSON, error) {
	p := filepath.Join(root, "package.json")
	bytes, err := os.ReadFile(p)

	if err != nil {
		return nil, err
	}

	var pkg PackageJSON
	err = json.Unmarshal(bytes, &pkg)

	return &pkg, err
}
