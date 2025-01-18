package notes

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
