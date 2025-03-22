package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"portal/quark/neutron/config"
	"portal/quark/neutron/services/filesystem"
)

type FilePorter struct {
	targetRootPath string
}

func NewFilePorter() (*FilePorter, error) {

	sourceUrl, ok := config.GetConfiguration("STORAGE_URL")
	if !ok || sourceUrl == nil {
		logrus.Fatalln("STORAGE_URL 未配置")
	}
	resolvedPath, err := filesystem.ResolvePath(sourceUrl.(string))
	if err != nil {
		logrus.Fatalln("NewFilePorter解析路径失败", err)
		return nil, fmt.Errorf("NewFilePorter解析路径失败")
	}
	return &FilePorter{targetRootPath: resolvedPath}, nil
}

func (p *FilePorter) CopyFile(srcPath, targetPath string) (string, error) {
	targetDir := filepath.Dir(targetPath)
	fullTargetDir := p.targetRootPath + "/" + targetDir
	err := os.MkdirAll(fullTargetDir, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("MkdirAll: %w", err)
	}
	fullTargetPath := p.targetRootPath + "/" + targetPath
	err = filesystem.CopyFile(srcPath, fullTargetPath)
	if err != nil {
		return "", fmt.Errorf("CopyFile: %w", err)
	}

	return fullTargetDir, nil

}
