package storage

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"portal/services/base58"
	filesystem2 "portal/services/filesystem"

	"github.com/pnnh/neutron/config"
	"github.com/pnnh/neutron/helpers"
	"github.com/pnnh/neutron/helpers/jsonmap"
	nemodels "github.com/pnnh/neutron/models"
	"github.com/pnnh/neutron/services/checksum"
	"github.com/pnnh/neutron/services/filesystem"

	"github.com/gin-gonic/gin"
)

func getFile(fullPath string) (*jsonmap.JsonMap, error) {
	fileStat, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("getFile error: %v ", err)
	}
	fileName := fileStat.Name()
	fileUid := ""
	mimeType := ""
	if fileStat.IsDir() {
		fileUid = base58.EncodeBase58String(fileName)
		mimeType = "directory"
	} else {
		sumValue, err := checksum.CalcSha256(fullPath)
		if err != nil {
			return nil, fmt.Errorf("计算文件校验和失败: %w", err)
		}
		fileUid = sumValue
		mimeType = helpers.GetMimeType(fileName)
	}
	if fileUid == "" {
		return nil, fmt.Errorf("getFile fileUid is Empty")
	}

	dataRow := jsonmap.NewJsonMap()
	dataRow.SetString("title", fileName)
	dataRow.SetString("uid", fileUid)

	portalUrl, ok := config.GetConfigurationString("PUBLIC_PORTAL_URL")
	if !ok {
		return nil, fmt.Errorf("getFile portalUrl is Empty")
	}
	pathHash := base58.EncodeBase58String(fullPath)
	fileUrl := fmt.Sprintf("%s/host/storage/files/data/%s", portalUrl, pathHash)
	dataRow.SetString("url", fileUrl)
	dataRow.SetString("mimetype", mimeType)
	isTextFile, err := filesystem2.IsTextFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("判断文件类型失败: %w", err)
	}
	innerMap := dataRow.InnerMap()
	innerMap["is_text"] = isTextFile
	innerMap["is_dir"] = fileStat.IsDir()
	innerMap["is_image"] = helpers.IsImageFile(fileName)
	innerMap["path"] = pathHash
	return dataRow, nil
}

func listFiles(targetDir string, showIgnore bool) ([]*jsonmap.JsonMap, error) {
	dir := targetDir

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("Error reading directory %q: %v\n", dir, err)
	}

	var noteFiles []*jsonmap.JsonMap
	for _, entry := range entries {
		// 获取名称
		fileName := entry.Name()

		isHidden, err := filesystem.IsHidden(fileName)
		if err != nil {
			return nil, fmt.Errorf("Error checking file %q: %v\n", fileName, err)
		}
		if isHidden && !showIgnore {
			continue
		}
		isIgnore := filesystem.IsIgnoredPath(fileName)
		if isIgnore && !showIgnore {
			continue
		}

		fullPath := filepath.Join(dir, fileName)
		if fullPath == "" {
			continue
		}
		dataRow, err := getFile(fullPath)
		if err != nil {
			return nil, fmt.Errorf("获取文件信息失败: %w", err)
		}
		innerMap := dataRow.InnerMap()
		innerMap["is_ignore"] = isIgnore

		noteFiles = append(noteFiles, dataRow)

	}
	return noteFiles, nil
}

// 在主机模式下遍历某个目录下的笔记列表
func HostFileSelectHandler(gctx *gin.Context) {
	dirParam := gctx.Query("dir")
	if dirParam == "" {
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(fmt.Errorf("dir参数不能为空"), "查询笔记出错"))
		return
	}
	dirData, err := base58.DecodeBase58String(dirParam)
	if err != nil {
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "查询笔记出错2"))
		return
	}
	showIgnore := gctx.Query("showIgnore")
	showIgnoreBoolean := false
	if showIgnore == "true" {
		showIgnoreBoolean = true
	}

	targetDir := string(dirData)
	selectResult, err := listFiles(targetDir, showIgnoreBoolean)
	if err != nil {
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "查询笔记出错2"))
		return
	}

	rangeMap := make([]map[string]interface{}, 0)
	for _, v := range selectResult {
		rangeMap = append(rangeMap, v.InnerMap())
	}

	resp := map[string]any{
		"page":  1,
		"size":  1,
		"count": 1,
		"range": rangeMap,
	}

	responseResult := nemodels.NECodeOk.WithData(resp)

	gctx.JSON(http.StatusOK, responseResult)
}

func HostFileDescHandler(gctx *gin.Context) {
	dirParam := gctx.Query("uid")
	if dirParam == "" {
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(fmt.Errorf("dir参数不能为空"), "查询笔记出错"))
		return
	}
	targetFile, err := base58.DecodeBase58String(dirParam)
	if err != nil {
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "查询笔记出错2"))
		return
	}

	fullPath, err := filesystem.ResolvePath(targetFile)
	if err != nil {
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "查询笔记出错1"))
		return
	}

	dataRow, err := getFile(fullPath)
	if err != nil {
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "查询笔记出错1"))
		return
	}

	responseResult := nemodels.NECodeOk.WithData(dataRow.InnerMap())

	gctx.JSON(http.StatusOK, responseResult)
}

func HostFileDataHandler(gctx *gin.Context) {
	dirParam := gctx.Param("uid")
	if dirParam == "" {
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(fmt.Errorf("dir参数不能为空"), "查询笔记出错"))
		return
	}
	targetFile, err := base58.DecodeBase58String(dirParam)
	if err != nil {
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "查询笔记出错2"))
		return
	}

	fullPath, err := filesystem.ResolvePath(targetFile)
	if err != nil {
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "查询笔记出错1"))
		return
	}
	gctx.File(fullPath)
}
