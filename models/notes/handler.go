package notes

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"portal/models"
)

func NoteSelectHandler(gctx *gin.Context) {
	selectResult, err := SelectNotes(1, 60)
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

func NoteViewerInsertHandler(gctx *gin.Context) {
	result := models.CodeOk.WithData(map[string]any{
		"changes": 1,
	})

	gctx.JSON(http.StatusOK, result)
}
