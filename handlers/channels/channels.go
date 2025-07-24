package channels

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"portal/business"
	"portal/business/channels"
	"portal/models"
)

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
		lang = business.DefaultLanguage
	}
	selectResult, err := channels.SelectChannels(keyword, pageInt, sizeInt, lang)
	if err != nil {
		gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "查询频道出错"))
		return
	}

	selectResponse := models.SelectResultToResponse(selectResult)
	responseResult := models.CodeOk.WithData(selectResponse)

	gctx.JSON(http.StatusOK, responseResult)
}

// 输入时自动完成
func ChannelCompleteHandler(gctx *gin.Context) {
	keyword := gctx.Query("keyword")
	lang := gctx.Query("lang")
	selectResult, err := channels.CompleteChannels(keyword, lang)
	if err != nil {
		gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "查询频道出错"))
		return
	}

	responseResult := models.CodeOk.WithData(selectResult)

	gctx.JSON(http.StatusOK, responseResult)
}

func ChannelGetHandler(gctx *gin.Context) {
	uid := gctx.Param("uid")
	lang := gctx.Query("lang")
	if uid == "" {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("uid不能为空"))
		return
	}
	selectResult, err := channels.PGGetChannel(uid, lang)
	if err != nil {
		gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "查询频道出错"))
		return
	}
	var modelData any
	if selectResult != nil {
		modelData = selectResult.ToViewModel()
	}
	responseResult := models.CodeOk.WithData(modelData)

	gctx.JSON(http.StatusOK, responseResult)
}
