package notes

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"portal/models"
	"portal/neutron/helpers"
)

func NoteSelectHandler(gctx *gin.Context) {
	keyword := gctx.Query("keyword")
	page := gctx.Query("page")
	size := gctx.Query("size")
	channel := gctx.Query("channel")
	pageInt, err := strconv.Atoi(page)
	if err != nil {
		pageInt = 1
	}
	sizeInt, err := strconv.Atoi(size)
	if err != nil {
		sizeInt = 10
	}
	selectResult, err := SelectNotes(channel, keyword, pageInt, sizeInt)
	if err != nil {
		gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "查询笔记出错"))
		return
	}
	responseResult := models.CodeOk.WithData(selectResult)

	gctx.JSON(http.StatusOK, responseResult)
}

func NoteGetHandler(gctx *gin.Context) {
	uid := gctx.Param("uid")
	if uid == "" {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("uid不能为空"))
		return
	}
	selectResult, err := PGGetNote(uid)
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
