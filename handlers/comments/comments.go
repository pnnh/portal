package comments

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"portal/models"
	"portal/neutron/server/helpers"
)

func CommentInsertHandler(gctx *gin.Context) {
	model := &models.CommentModel{}
	if err := gctx.ShouldBindJSON(model); err != nil {
		gctx.JSON(http.StatusOK, models.CodeError.WithError(err))
		return
	}

	accountModel := &models.AccountModel{
		Urn:         helpers.MustUuid(),
		Username:    model.EMail,
		Password:    "",
		Photo:       "",
		CreateTime:  time.Now().UTC(),
		UpdateTime:  time.Now().UTC(),
		Nickname:    "",
		EMail:       model.EMail,
		Credentials: "",
		Session:     "",
		Description: "",
		Status:      0,
		Website:     model.Website,
		Fingerprint: model.Fingerprint,
	}

	err := models.EnsureAccount(accountModel)
	if err != nil {
		gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "更新用户账户出错"))
		return
	}

	model.Urn = helpers.MustUuid()
	model.CreateTime = time.Now().UTC()
	model.UpdateTime = time.Now().UTC()
	model.Creator = accountModel.Urn
	model.Thread = helpers.MustUuid()
	model.Referer = helpers.MustUuid()
	model.Resource = helpers.MustUuid()
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
