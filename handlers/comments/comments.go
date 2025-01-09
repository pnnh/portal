package comments

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"portal/business"
	"portal/models"
	"portal/neutron/helpers"
)

type CommentInsertRequest struct {
	models.CommentModel
}

func CommentInsertHandler(gctx *gin.Context) {
	model := &CommentInsertRequest{}
	if err := gctx.ShouldBindJSON(model); err != nil {
		gctx.JSON(http.StatusOK, models.CodeError.WithError(err))
		return
	}

	accountModel, err := business.FindUserFromCookie(gctx)
	if err != nil {
		logrus.Println("GetAccountBySessionId", err)
		gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "查询账号出错"))
		return
	}
	if accountModel == nil {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("账号不存在"))
		return
	}

	model.Urn = helpers.MustUuid()
	model.CreateTime = time.Now().UTC()
	model.UpdateTime = time.Now().UTC()
	model.Creator = accountModel.Urn
	model.Thread = helpers.EmptyUuid()
	model.Referer = helpers.EmptyUuid()
	model.IPAddress = helpers.GetIpAddress(gctx)
	model.EMail = accountModel.EMail
	model.Nickname = accountModel.Nickname
	model.Website = accountModel.Website

	err = models.PGInsertComment(&model.CommentModel)
	if err != nil {
		gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "插入评论出错"))
		return
	}

	result := models.CodeOk.WithData(map[string]any{
		"changes": 1,
		"urn":     model.Urn,
	})

	gctx.JSON(http.StatusOK, result)
}

func CommentSelectHandler(gctx *gin.Context) {
	target := gctx.Query("resource")
	if target == "" {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("资源不存在"))
		return
	}

	selectResult, err := models.SelectComments(target, 1, 60)
	if err != nil {
		gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "查询评论出错"))
		return
	}
	responseResult := models.CodeOk.WithData(selectResult)

	gctx.JSON(http.StatusOK, responseResult)
}
