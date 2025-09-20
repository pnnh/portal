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
	ignoreArray = append(ignoreArray, ".ds_store")
	ignoreArray = append(ignoreArray, "__pycache__")
	ignoreArray = append(ignoreArray, ".pytest_cache")
	ignoreArray = append(ignoreArray, ".mypy_cache")
	ignoreArray = append(ignoreArray, "go"+string(filepath.Separator)+"src")
	ignoreArray = append(ignoreArray, "go"+string(filepath.Separator)+"pkg")
	ignoreArray = append(ignoreArray, ".imagelibrary")
	ignoreArray = append(ignoreArray, ".imagechannel")
	ignoreArray = append(ignoreArray, ".notelibrary")
	ignoreArray = append(ignoreArray, "vcpkg_installed")
	ignoreArray = append(ignoreArray, "x64-windows")
	ignoreArray = append(ignoreArray, "x86-windows")
	ignoreArray = append(ignoreArray, "generated files")
	ignoreArray = append(ignoreArray, ".sass-cache")
	ignoreArray = append(ignoreArray, "temp")
	ignoreArray = append(ignoreArray, ".bundle")
	ignoreArray = append(ignoreArray, ".cache")
	ignoreArray = append(ignoreArray, ".config")
	ignoreArray = append(ignoreArray, ".local")
	ignoreArray = append(ignoreArray, ".npm")
	ignoreArray = append(ignoreArray, ".nuget")
	ignoreArray = append(ignoreArray, ".parcel-cache")
	ignoreArray = append(ignoreArray, ".pub-cache")
	ignoreArray = append(ignoreArray, "Pods")
	ignoreArray = append(ignoreArray, "Carthage")
	ignoreArray = append(ignoreArray, "builds")
	ignoreArray = append(ignoreArray, "DerivedData")
	ignoreArray = append(ignoreArray, "xcuserdata")
	ignoreArray = append(ignoreArray, "fastlane")
	ignoreArray = append(ignoreArray, "node_modules")
	ignoreArray = append(ignoreArray, "jspm_packages")
	ignoreArray = append(ignoreArray, "coverage")
	ignoreArray = append(ignoreArray, "cache")
	ignoreArray = append(ignoreArray, "bin")
	ignoreArray = append(ignoreArray, "obj")
	ignoreArray = append(ignoreArray, "logs")
	ignoreArray = append(ignoreArray, "log")
	ignoreArray = append(ignoreArray, "debug")
	ignoreArray = append(ignoreArray, "release")
	ignoreArray = append(ignoreArray, "tmp")
	ignoreArray = append(ignoreArray, "temp")
	ignoreArray = append(ignoreArray, "Pods")

	lowerFullPath := string(filepath.Separator) + strings.ToLower(fullPath) + string(filepath.Separator)
	for _, ignoreItem := range ignoreArray {
		if strings.Contains(lowerFullPath, string(filepath.Separator)+strings.ToLower(ignoreItem)+string(filepath.Separator)) {
			return true
		}
	}
	return false
}
