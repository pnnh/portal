package filesystem

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func ResolvePath(path string) (string, error) {

	resolvedPath := path

	if strings.HasPrefix(path, "file://") {
		resolvedPath = strings.Replace(path, "file://", "", -1)
	}

	if strings.HasPrefix(resolvedPath, "./") {
		dir, err := os.Getwd()
		if err != nil {
			return path, fmt.Errorf("获取当前目录失败: %s", err)
		}
		resolvedPath = strings.Replace(resolvedPath, "./", dir, 1)
	}

	if strings.HasPrefix(resolvedPath, "~/") {
		userHomeDir, err := os.UserHomeDir()
		if err != nil {
			return path, fmt.Errorf("获取用户目录失败: %s", err)
		}
		resolvedPath = strings.Replace(resolvedPath, "~/", userHomeDir+"/", 1)
	}

	return resolvedPath, nil
}

func CopyFile(src, dst string) (err error) {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer func(sourceFile *os.File) {
		err = sourceFile.Close()
	}(sourceFile)

	destinationFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer func(destinationFile *os.File) {
		err = destinationFile.Close()
	}(destinationFile)

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	err = destinationFile.Sync()
	if err != nil {
		return fmt.Errorf("failed to sync destination file: %w", err)
	}

	return nil
}
