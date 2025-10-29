package images

import (
	"net/http"
	"strconv"

	nemodels "neutron/models"

	"github.com/gin-gonic/gin"
)

func ImageSelectHandler(gctx *gin.Context) {
	keyword := gctx.Query("keyword")
	page := gctx.Query("page")
	size := gctx.Query("size")
	pageInt, err := strconv.Atoi(page)
	if err != nil {
		pageInt = 1
	}
	sizeInt, err := strconv.Atoi(size)
	if err != nil {
		sizeInt = 20
	}
	selectResult, err := SelectImages(keyword, pageInt, sizeInt)
	if err != nil {
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "查询图片出错"))
		return
	}

	selectResponse := nemodels.NESelectResultToResponse(selectResult)
	responseResult := nemodels.NECodeOk.WithData(selectResponse)

	gctx.JSON(http.StatusOK, responseResult)
}

func ImageGetHandler(gctx *gin.Context) {
	uid := gctx.Param("uid")
	if uid == "" {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("uid不能为空"))
		return
	}
	selectResult, err := PGGetImage(uid)
	if err != nil {
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "查询图片出错"))
		return
	}
	responseResult := nemodels.NECodeOk.WithData(selectResult)

	gctx.JSON(http.StatusOK, responseResult)
}
