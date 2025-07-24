package notes

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"neutron/helpers"
	"portal/business"
	"portal/business/channels"
	"portal/models"
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
		lang = business.DefaultLanguage
	}
	selectResult, err := SelectNotes(channel, keyword, pageInt, sizeInt, lang)
	if err != nil {
		gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "查询笔记出错"))
		return
	}

	selectResponse := models.SelectResultToResponse(selectResult)
	responseResult := models.CodeOk.WithData(selectResponse)

	gctx.JSON(http.StatusOK, responseResult)
}

func NoteInsertHandler(gctx *gin.Context) {
	accountModel, err := business.FindAccountFromCookie(gctx)
	if err != nil {
		logrus.Warnln("NoteInsertHandler", err)
		gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "查询账号出错c"))
		return
	}
	if accountModel == nil || accountModel.IsAnonymous() {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("账号不存在或匿名用户不能发布笔记"))
		return
	}

	model := &MTNoteModel{}
	if err := gctx.ShouldBindJSON(model); err != nil {
		gctx.JSON(http.StatusOK, models.CodeError.WithError(err))
		return
	}
	if model.Title == "" || model.Body == "" {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("标题或内容不能为空"))
		return
	}
	if model.Lang == "" || (model.Lang != business.LangZh && model.Lang != business.LangEn) {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("Lang参数错误"))
		return
	}
	if model.Channel == "" {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("Channel参数不能为空"))
		return
	}

	model.Uid = helpers.MustUuid()
	model.Owner = accountModel.Uid
	model.CreateTime = time.Now().UTC()
	model.UpdateTime = time.Now().UTC()
	model.Status = 0 // 待审核
	model.Cid = model.Uid
	model.Dc = business.CurrentDC

	channelModel, err := channels.PGGetChannel(model.Channel, model.Lang)
	if err != nil {
		gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "查询频道出错"))
		return
	}
	if channelModel == nil {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("频道不存在"))
		return
	}

	err = PGInsertNote(model)
	if err != nil {
		gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "插入笔记出错"))
		return
	}

	result := models.CodeOk.WithData(map[string]any{
		"changes": 1,
		"uid":     model.Uid,
	})

	gctx.JSON(http.StatusOK, result)
}

func NoteUpdateHandler(gctx *gin.Context) {
	uid := gctx.Param("uid")
	if uid == "" {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("uid不能为空"))
		return
	}
	accountModel, err := business.FindAccountFromCookie(gctx)
	if err != nil {
		logrus.Warnln("NoteUpdateHandler", err)
		gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "查询账号出错c"))
		return
	}
	if accountModel == nil || accountModel.IsAnonymous() {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("账号不存在或匿名用户不能修改笔记"))
		return
	}

	model := &MTNoteModel{}
	if err := gctx.ShouldBindJSON(model); err != nil {
		gctx.JSON(http.StatusOK, models.CodeError.WithError(err))
		return
	}
	if model.Title == "" || model.Body == "" {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("标题或内容不能为空"))
		return
	}
	oldModel, err := PGGetNote(uid, "")
	if err != nil {
		gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "查询笔记出错"))
		return
	}
	if oldModel == nil {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("笔记不存在"))
		return
	}
	if oldModel.Owner != accountModel.Uid {
		gctx.JSON(http.StatusOK, models.CodeUnauthorized.WithMessage("没有权限修改该笔记"))
		return
	}

	model.Uid = uid
	model.UpdateTime = time.Now().UTC()

	err = PGUpdateNote(model)
	if err != nil {
		gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "更新笔记出错"))
		return
	}

	result := models.CodeOk.WithData(model.Uid)

	gctx.JSON(http.StatusOK, result)
}

func NoteGetHandler(gctx *gin.Context) {
	uid := gctx.Param("uid")
	lang := gctx.Query("lang")
	if uid == "" {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("uid不能为空"))
		return
	}
	selectResult, err := PGGetNote(uid, lang)
	if err != nil {
		gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "查询笔记出错"))
		return
	}
	responseResult := models.CodeOk.WithData(selectResult)

	gctx.JSON(http.StatusOK, responseResult)
}

type ViewerInsertRequest struct {
	ClientIp string `json:"clientIp"`
}

func NoteViewerInsertHandler(gctx *gin.Context) {
	uid := gctx.Param("uid")
	if uid == "" {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("uid不能为空"))
		return
	}
	request := &ViewerInsertRequest{}
	if err := gctx.ShouldBindJSON(request); err != nil {
		gctx.JSON(http.StatusOK, models.CodeError.WithError(err))
		return
	}

	model := &MTViewerModel{
		Uid:        helpers.MustUuid(),
		Target:     uid,
		Address:    request.ClientIp,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
		Class:      "note",
	}
	opErr, itemErrs := PGInsertViewer(model)
	if opErr != nil {
		gctx.JSON(http.StatusOK, models.CodeError.WithError(opErr))
		return
	}
	for key, item := range itemErrs {
		if !errors.Is(item, ErrViewerLogExists) {
			logrus.Warnln("NoteViewerInsertHandler", key, item)
		}
	}
	result := models.CodeOk.WithData(map[string]any{
		"changes": 1,
	})

	gctx.JSON(http.StatusOK, result)
}
