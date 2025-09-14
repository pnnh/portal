package channels

import (
	"net/http"
	nemodels "neutron/models"
	"strconv"

	"github.com/gin-gonic/gin"
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

func ChannelSelectHandler(gctx *gin.Context) {
	keyword := gctx.Query("keyword")
	page := gctx.Query("page")
	size := gctx.Query("size")
	lang := gctx.Query("lang")
	pageInt, err := strconv.Atoi(page)
	if err != nil {
		pageInt = 1
	}
	sizeInt, err := strconv.Atoi(size)
	if err != nil {
		sizeInt = 10
	}
	if lang == "" {
		lang = nemodels.DefaultLanguage
	}
	selectResult, err := SelectChannels(keyword, pageInt, sizeInt, lang)
	if err != nil {
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "查询频道出错"))
		return
	}

	selectResponse := nemodels.NESelectResultToResponse(selectResult)
	responseResult := nemodels.NECodeOk.WithData(selectResponse)

	gctx.JSON(http.StatusOK, responseResult)
}

// 输入时自动完成
func ChannelCompleteHandler(gctx *gin.Context) {
	keyword := gctx.Query("keyword")
	lang := gctx.Query("lang")
	selectResult, err := PGCompleteChannels(keyword, lang)
	if err != nil {
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "查询频道出错"))
		return
	}

	responseResult := nemodels.NECodeOk.WithData(selectResult)

	gctx.JSON(http.StatusOK, responseResult)
}

//
//func ChannelGetByCidHandler(gctx *gin.Context) {
//	cid := gctx.Param("cid")
//	lang := gctx.Param("lang")
//	wangLang := gctx.Param("wantLang")
//	if lang == "" {
//		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithLocalMessage(nemodels.LangEn,
//			"lang不能为空", "lang cannot be empty"))
//		return
//	}
//	if cid == "" || wangLang == "" {
//		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithLocalMessage(lang,
//			"cid或wantLang不能为空", "cid or wantLang cannot be empty"))
//		return
//	}
//
//	selectResult, err := PGGetChannelByCid(cid, wangLang)
//	if err != nil {
//		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithLocalError(lang, err, "查询频道出错", "query channel failed"))
//		return
//	}
//	if selectResult == nil {
//		gctx.JSON(http.StatusOK, nemodels.NECodeNotFound.WithLocalMessage(lang, "频道不存在", "channel not found"))
//		return
//	}
//	var modelData = selectResult.ToViewModel()
//
//	responseResult := nemodels.NECodeOk.WithData(modelData)
//
//	gctx.JSON(http.StatusOK, responseResult)
//}
