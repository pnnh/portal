package notes

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	nemodels "neutron/models"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"neutron/config"
	"neutron/services/filesystem"
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

	mtNote, err := PGGetNote(uid, "")
	if err != nil || mtNote == nil {
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
	if mtNote.RepoId == "" || mtNote.Branch == "" {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("RepoId或Branch为空"))
		return
	}
	assetsPath := fmt.Sprintf("%s/%s/%s/%s/%s", storageUrl, mtNote.RepoId, mtNote.Branch,
		strings.TrimLeft(mtNote.RepoPath, "/"), decodedParent)
	fullAssetsPath, err := filesystem.ResolvePath(assetsPath)
	if err != nil {
		log.Println("ResolvePath", err)
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "ResolvePath出错"))
		return
	}
	log.Println("fullAssetsPath", fullAssetsPath)

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
