package notes

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	nemodels "neutron/models"

	"neutron/config"
	"neutron/services/filesystem"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func NoteAssetsSelectHandler(gctx *gin.Context) {
	uid := gctx.Param("uid")
	if uid == "" {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("uid不能为空"))
		return
	}
	parent := gctx.Query("parent")
	decodedParent := ""
	if parent != "" {
		decodeString, err := base64.URLEncoding.DecodeString(parent)
		if err != nil {
			gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "DecodeString parent出错"))
			return
		}
		decodedParent = string(decodeString)
	}

	noteTable, err := PGGetNote(uid, "")
	if err != nil || noteTable == nil {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithError(err))
		return
	}

	storageUrl, ok := config.GetConfigurationString("STORAGE_URL")
	if !ok || storageUrl == "" {
		//return fmt.Errorf("STORAGE_URL 未配置2")
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("STORAGE_URL 未配置2"))
		return
	}
	storagePath, err := filesystem.ResolvePath(storageUrl)
	if err != nil {
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "ResolvePath出错"))
		return
	}
	noteModel := noteTable.ToModel()
	if noteModel.RepoFirstCommit == "" || noteModel.Branch == "" {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("RepoFirstCommit或Branch为空"))
		return
	}
	assetsPath := fmt.Sprintf("%s/%s/%s/%s/%s", storageUrl, noteTable.RepoFirstCommit.String, noteTable.Branch.String,
		strings.TrimLeft(filepath.Dir(noteModel.RelativePath), "/"), decodedParent)
	assetsPath = strings.TrimRight(assetsPath, "/")
	fullAssetsPath, err := filesystem.ResolvePath(assetsPath)
	if err != nil {
		logrus.Println("ResolvePath", err)
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "ResolvePath出错"))
		return
	}
	logrus.Println("fullAssetsPath", fullAssetsPath)

	fileList, err := listFirstLevel(storagePath, fullAssetsPath)
	if err != nil {
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "listFirstLevel出错"))
		return
	}
	selectData := &nemodels.NESelectResponse{
		Page:  1,
		Size:  999,
		Count: len(fileList),
		Range: fileList,
	}
	responseResult := nemodels.NECodeOk.WithData(selectData)

	gctx.JSON(http.StatusOK, responseResult)
}

func listFirstLevel(storagePath, currentDir string) ([]any, error) {
	fileList := make([]any, 0)
	entries, err := os.ReadDir(currentDir)
	if err != nil {
		return nil, fmt.Errorf("os.ReadDir: %v", err)
	}

	for _, entry := range entries {
		if filesystem.IsExcludedFile(entry.Name()) {
			continue
		}
		relativePath := entry.Name()
		extName := strings.Trim(strings.ToLower(filepath.Ext(relativePath)), " ")
		model := MTNoteFileModel{
			Title:       relativePath,
			Path:        relativePath,
			IsDir:       entry.IsDir(),
			IsText:      filesystem.IsTextFile(extName),
			IsImage:     filesystem.IsImageFile(extName),
			StoragePath: strings.TrimPrefix(currentDir, storagePath) + relativePath,
		}
		fileList = append(fileList, model)
	}
	return fileList, nil
}
