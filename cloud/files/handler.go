package files

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"portal/business"

	"github.com/pnnh/neutron/config"
	nemodels "github.com/pnnh/neutron/models"
	"github.com/pnnh/neutron/services/datastore"
	"github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

func fileGetOutView(storageUrl string, dataRow *datastore.DataRow) (map[string]interface{}, error) {
	outView := make(map[string]interface{})
	outView["uid"] = dataRow.GetString("uid")
	outView["title"] = dataRow.GetString("title")
	outView["header"] = dataRow.GetStringOrEmpty("header")
	outView["body"] = dataRow.GetStringOrEmpty("body")
	outView["description"] = dataRow.GetStringOrEmpty("description")
	outView["keywords"] = dataRow.GetStringOrDefault("keywords", "")
	outView["status"] = dataRow.GetInt("status")
	outView["cover"] = dataRow.GetStringOrDefault("cover", "")
	outView["owner"] = dataRow.GetNullString("owner")
	outView["discover"] = dataRow.GetInt("discover")
	outView["partition"] = dataRow.GetStringOrDefault("partition", "")
	outView["create_time"] = dataRow.GetTime("create_time")
	outView["update_time"] = dataRow.GetTime("update_time")
	outView["version"] = dataRow.GetStringOrDefault("version", "")
	outView["build"] = dataRow.GetStringOrDefault("build", "")
	fileUrl := dataRow.GetStringOrDefault("url", "")
	outView["url"] = strings.Replace(fileUrl, "storage://", storageUrl, 1)
	outView["branch"] = dataRow.GetStringOrDefault("branch", "")
	outView["commit"] = dataRow.GetStringOrDefault("commit", "")
	outView["name"] = dataRow.GetStringOrDefault("name", "")
	outView["checksum"] = dataRow.GetStringOrDefault("checksum", "")
	outView["syncno"] = dataRow.GetStringOrDefault("syncno", "")
	outView["channel_name"] = dataRow.GetStringOrDefault("channel_name", "")
	outView["mimetype"] = dataRow.GetStringOrDefault("mimetype", "")
	outView["object_uid"] = dataRow.GetStringOrDefault("object_uid", "")

	return outView, nil
}

// 在主机模式下遍历某个目录下的笔记列表
func CloudFileSelectHandler(gctx *gin.Context) {
	keyword := gctx.Query("keyword")
	page := gctx.Query("page")
	size := gctx.Query("size")
	parent := gctx.Query("parent")
	pageInt, err := strconv.Atoi(page)
	if err != nil {
		logrus.Warningln("pageInt warning", err)
		pageInt = 1
	}
	sizeInt, err := strconv.Atoi(size)
	if err != nil {
		sizeInt = 10
	}
	accountModel, err := business.FindAccountFromCookie(gctx)
	if err != nil {
		logrus.Warnln("ConsoleNotesSelectHandler", err)
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "查询账号出错b"))
		return
	}
	if accountModel == nil {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("账号不存在"))
		return
	}
	viewParam := gctx.Query("viewType")
	if viewParam != "filesystem" && viewParam != "library" {
		viewParam = "library"
	}
	selectParams := &FileSelectParams{
		Parent: parent,
	}
	pagination, selectResult, err := SelectFiles(keyword, pageInt, sizeInt, selectParams)
	if err != nil {
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "查询笔记出错2"))
		return
	}

	portalUrl, ok := config.GetConfigurationString("PUBLIC_PORTAL_URL")
	if !ok || portalUrl == "" {
		logrus.Warnln("PUBLIC_PORTAL_URL 未配置2")
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "PUBLIC_PORTAL_URL 未配置3"))
		return
	}
	storageUrl := portalUrl + "/storage/"
	respView := make([]map[string]interface{}, 0)
	for _, v := range selectResult {
		outView, err := fileGetOutView(storageUrl, v)
		if err != nil {
			gctx.JSON(http.StatusOK, nemodels.NECodeError.WithError(err))
			return
		}
		respView = append(respView, outView)
	}

	resp := map[string]any{
		"page":  pagination.Page,
		"size":  pagination.Size,
		"count": pagination.Count,
		"range": respView,
	}

	responseResult := nemodels.NECodeOk.WithData(resp)

	gctx.JSON(http.StatusOK, responseResult)
}

func CloudFileDescHandler(gctx *gin.Context) {
	uid := gctx.Query("uid")
	if uid == "" {
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(fmt.Errorf("dir参数不能为空"), "查询笔记出错"))
		return
	}
	accountModel, err := business.FindAccountFromCookie(gctx)
	if err != nil {
		logrus.Warnln("ConsoleNotesSelectHandler", err)
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "查询账号出错b"))
		return
	}
	if accountModel == nil {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("账号不存在"))
		return
	}
	dataRow, err := PGGetFile(accountModel.Uid, uid)
	if err != nil {
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "查询笔记出错1"))
		return
	}

	responseResult := nemodels.NECodeOk.WithData(dataRow.InnerMap())

	gctx.JSON(http.StatusOK, responseResult)
}
