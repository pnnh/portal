package comments

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"portal/models"
	"portal/neutron/helpers"
)

func CommentInsertHandler(gctx *gin.Context) {
	model := &models.CommentModel{}
	if err := gctx.ShouldBindJSON(model); err != nil {
		gctx.JSON(http.StatusOK, models.CodeError.WithError(err))
		return
	}

	accountModel, err := models.EnsureAccount(model.EMail, model.Nickname, model.EMail, model.Website, "", model.Fingerprint)
	if err != nil {
		gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "更新用户账户出错"))
		return
	}
	if accountModel != nil {
		model.Creator = accountModel.Urn
	}

	model.Urn = helpers.MustUuid()
	model.CreateTime = time.Now().UTC()
	model.UpdateTime = time.Now().UTC()
	model.Creator = helpers.EmptyUuid()
	model.Thread = helpers.EmptyUuid()
	model.Referer = helpers.EmptyUuid()
	model.Resource = helpers.EmptyUuid()
	model.IPAddress = helpers.GetIpAddress(gctx)

	err = models.PGInsertComment(model)
	if err != nil {
		gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "插入评论出错"))
		return
	}

	result := models.CodeOk.WithData(map[string]any{})

	gctx.JSON(http.StatusOK, result)
}

func CommentSelectHandler(gctx *gin.Context) {
	selectResult, err := models.SelectComments(1, 30)
	if err != nil {
		gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "查询评论出错"))
		return
	}
	responseResult := models.CodeOk.WithData(selectResult)

	gctx.JSON(http.StatusOK, responseResult)
}
