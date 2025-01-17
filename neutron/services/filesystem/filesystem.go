package filesystem

import (
	"fmt"
	"os"
	"strings"
)

func ResolvePath(path string) (string, error) {

	resolvedPath := path

	if strings.HasPrefix(path, "file://") {
		resolvedPath = strings.Replace(path, "file://", "", -1)
	}

	if strings.HasPrefix(resolvedPath, "./") {
		resolvedPath = strings.Replace(resolvedPath, "./", "", 1)
	}

	if strings.HasPrefix(resolvedPath, "~/") {
		userHomeDir, err := os.UserHomeDir()
		if err != nil {
			return path, fmt.Errorf("获取用户目录失败: %s", err)
		}
		resolvedPath = strings.Replace(resolvedPath, "/", userHomeDir+"/", 1)
	}

	return resolvedPath, nil
}
