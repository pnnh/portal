package notes

import (
	"net/http"
	"strconv"
	"time"

	nemodels "neutron/models"

	"neutron/helpers"
	"portal/business"
	"portal/business/channels"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func NoteSelectHandler(gctx *gin.Context) {
	keyword := gctx.Query("keyword")
	page := gctx.Query("page")
	size := gctx.Query("size")
	channel := gctx.Query("channel")
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
	pagination, selectResult, err := SelectNotes(channel, keyword, pageInt, sizeInt, lang)
	if err != nil {
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "查询笔记出错"))
		return
	}

	resp := map[string]any{
		"page":  pagination.Page,
		"size":  pagination.Size,
		"count": pagination.Count,
		"range": selectResult,
	}

	//selectResponse := nemodels.NESelectResultToResponse(selectResult)
	responseResult := nemodels.NECodeOk.WithData(resp)

	gctx.JSON(http.StatusOK, responseResult)
}

func NoteConsoleInsertHandler(gctx *gin.Context) {
	accountModel, err := business.FindAccountFromCookie(gctx)
	if err != nil {
		logrus.Warnln("NoteConsoleInsertHandler", err)
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "查询账号出错c"))
		return
	}
	if accountModel == nil || accountModel.IsAnonymous() {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("账号不存在或匿名用户不能发布笔记"))
		return
	}

	model := &MTNoteView{}
	if err := gctx.ShouldBindJSON(model); err != nil {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithError(err))
		return
	}
	if model.Title == "" || model.Body == "" {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("标题或内容不能为空"))
		return
	}
	if model.Lang == "" || (model.Lang != nemodels.LangZh && model.Lang != nemodels.LangEn) {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("Lang参数错误"))
		return
	}
	if model.Channel == "" {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("Channel参数不能为空"))
		return
	}

	model.Uid = helpers.MustUuid()
	model.Owner = accountModel.Uid
	model.CreateTime = time.Now().UTC()
	model.UpdateTime = time.Now().UTC()
	model.Status = 0 // 待审核
	model.Dc = business.CurrentDC
	if model.Name == "" {
		model.Name = model.Uid
	}

	channelModel, err := channels.PGConsoleGetChannelByUid(model.Channel)
	if err != nil {
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "查询频道出错"))
		return
	}
	if channelModel == nil {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("频道不存在2"))
		return
	}

	err = PGConsoleInsertNote(model)
	if err != nil {
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "插入笔记出错"))
		return
	}

	result := nemodels.NECodeOk.WithData(map[string]any{
		"changes": 1,
		"uid":     model.Uid,
	})

	gctx.JSON(http.StatusOK, result)
}

func NoteGetHandler(gctx *gin.Context) {
	uid := gctx.Param("uid")
	lang := gctx.Query("lang")
	if uid == "" {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("uid不能为空"))
		return
	}
	noteTable, err := PGGetNote(uid, lang)
	if err != nil {
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "查询笔记出错"))
		return
	}
	noteModel := noteTable.ToModel()
	responseResult := nemodels.NECodeOk.WithLocalData(lang, noteModel)

	gctx.JSON(http.StatusOK, responseResult)
}
