package filesystem

import (
	"path/filepath"
)

func IsTextFile(fileName string) (bool, error) {
	extName := filepath.Ext(fileName)
	textFileExts := map[string]bool{
		".txt":  true,
		".md":   true,
		".json": true,
		".xml":  true,
		".yaml": true,
		".yml":  true,
		".csv":  true,
		".log":  true,
		".html": true,
		".htm":  true,
		".css":  true,
		".js":   true,
		".go":   true,
		".py":   true,
		".java": true,
		".c":    true,
		".cpp":  true,
		".h":    true,
		".sh":   true,
	}
	_, exists := textFileExts[extName]
	return exists, nil
}
