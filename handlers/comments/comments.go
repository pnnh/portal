package comments

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"portal/business"
	"portal/models"
	"portal/models/notes"
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

	accountModel, err := business.FindAccountFromCookie(gctx)
	if err != nil {
		logrus.Warnln("CommentInsertHandler", err)
		gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "查询账号出错c"))
		return
	}
	if accountModel == nil {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("账号不存在"))
		return
	}

	model.Uid = helpers.MustUuid()
	model.CreateTime = time.Now().UTC()
	model.UpdateTime = time.Now().UTC()
	model.Creator = accountModel.Uid
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
		"uid":     model.Uid,
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

	addr := helpers.GetIpAddress(gctx)
	commentViewers := make([]*notes.MTViewerModel, 0)
	for _, item := range selectResult.Range {
		comment := item.(*models.CommentModel)
		model := &notes.MTViewerModel{
			Uid:        helpers.MustUuid(),
			Target:     comment.Uid,
			Address:    addr,
			CreateTime: time.Now(),
			UpdateTime: time.Now(),
			Class:      "comment",
		}
		commentViewers = append(commentViewers, model)
	}

	opErr, itemErrs := notes.PGInsertViewer(commentViewers...)
	if opErr != nil {
		gctx.JSON(http.StatusOK, models.CodeError.WithError(opErr))
		return
	}
	for key, item := range itemErrs {
		if !errors.Is(item, notes.ErrViewerLogExists) {
			logrus.Warnln("CommentSelectHandler", key, item)
		}
	}

	gctx.JSON(http.StatusOK, responseResult)
}
