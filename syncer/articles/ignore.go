package articles

import (
	"path/filepath"
	"strings"
)

func IsIgnoredPath(fullPath string) bool {
	ignoreArray := make([]string, 0)
	ignoreArray = append(ignoreArray, ".git")
	ignoreArray = append(ignoreArray, ".idea")
	ignoreArray = append(ignoreArray, "node_modules")
	ignoreArray = append(ignoreArray, ".vscode")
	ignoreArray = append(ignoreArray, "vendor")
	ignoreArray = append(ignoreArray, "bin")
	ignoreArray = append(ignoreArray, "obj")
	ignoreArray = append(ignoreArray, "dist")
	ignoreArray = append(ignoreArray, "build")
	ignoreArray = append(ignoreArray, "out")
	ignoreArray = append(ignoreArray, ".DS_Store")
	ignoreArray = append(ignoreArray, "__pycache__")
	ignoreArray = append(ignoreArray, ".pytest_cache")
	ignoreArray = append(ignoreArray, ".mypy_cache")
	ignoreArray = append(ignoreArray, "go"+string(filepath.Separator)+"src")
	ignoreArray = append(ignoreArray, "go"+string(filepath.Separator)+"pkg")
	ignoreArray = append(ignoreArray, ".imagelibrary")
	ignoreArray = append(ignoreArray, ".imagechannel")
	ignoreArray = append(ignoreArray, ".notelibrary")
	for _, ignoreItem := range ignoreArray {
		if strings.Contains(fullPath, string(filepath.Separator)+ignoreItem+string(filepath.Separator)) {
			return true
		}
	}
	return false
}
