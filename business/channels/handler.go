package channels

import (
	"github.com/gin-gonic/gin"
	"net/http"
	nemodels "neutron/models"
)

func ChannelGetByUidHandler(gctx *gin.Context) {
	uid := gctx.Param("uid")
	lang := gctx.Param("lang")
	if lang == "" {
		lang = nemodels.DefaultLanguage
	}
	if uid == "" {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithLocalMessage(lang, "uid不能为空", "uid cannot be empty"))
		return
	}

	model := (&MTChannelModel{}).PGGetByUid(uid).ToModel()
	if model.Error != nil {
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(model.Error, "查询频道出错"))
		return
	}
	var modelData = model.ToViewModel()
	responseResult := nemodels.NECodeOk.WithData(modelData)

	gctx.JSON(http.StatusOK, responseResult)
}
