package images

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"portal/models/images"
	"portal/neutron/config"
	"portal/neutron/helpers"
	"portal/neutron/services/filesystem"
)

type SyncImagesWorker struct {
	repoPath   string
	filePorter *filesystem.FilePorter
}

func NewSyncImagesWorker(repoPath string) (*SyncImagesWorker, error) {
	storageUrl, ok := config.GetConfigurationString("STORAGE_URL")
	if !ok || storageUrl == "" {
		logrus.Fatalln("STORAGE_URL 未配置")
	}
	filePorter, err := filesystem.NewFilePorter(storageUrl)
	if err != nil {
		logrus.Errorln("初始化FilePorter失败", err)
		return nil, fmt.Errorf("初始化FilePorter失败")
	}
	return &SyncImagesWorker{
		repoPath:   repoPath,
		filePorter: filePorter,
	}, nil
}

func (w *SyncImagesWorker) StartWork() {
	err := filepath.Walk(w.repoPath, w.visitFile)
	if err != nil {
		logrus.Fatalln("error walking the path %v: %v\n", w.repoPath, err)
	}
}

func (w *SyncImagesWorker) visitFile(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if info.IsDir() {
		return nil
	}
	if !filesystem.IsImageFile(path) {
		return nil
	}
	if strings.Index(path, ".imagechannel") < 0 {
		return nil
	}
	relativePath, err := filepath.Rel(w.repoPath, path)
	if err != nil {
		return fmt.Errorf("filepath.Rel(%q, %q): %v", w.repoPath, path, err)
	}
	uid, err := helpers.StringToMD5UUID(path)
	if err != nil {
		return fmt.Errorf("helpers.StringToMD5UUID(%q): %v", path, err)
	}
	extName := filepath.Ext(relativePath)
	note := &images.MTImageModel{
		Uid:         uid,
		Title:       info.Name(),
		Description: info.Name(),
		Keywords:    "",
		Status:      1,
		Owner:       sql.NullString{},
		Channel:     sql.NullString{},
		Discover:    0,
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
		FilePath:    relativePath,
		ExtName:     extName,
	}
	err = images.PGInsertImage(note)
	if err != nil {
		return fmt.Errorf("插入图片失败: %w", err)
	}
	targetPath := fmt.Sprintf("images/%s%s", uid, extName)
	_, err = w.filePorter.CopyFile(path, targetPath)
	if err != nil {
		return fmt.Errorf("CopyFile: %w", err)
	}

	return nil
}
