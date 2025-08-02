package public

import (
	"net/http"
	"portal/models"

	"github.com/gin-gonic/gin"
)

func PublicApplicationSelectHandler(gctx *gin.Context) {
	accounts, err := models.SelectApplicationsByStatus(1)
	if err != nil {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithError(err))
		return
	}
	sessionData := map[string]interface{}{
		"list": accounts,
	}

	result := nemodels.NECodeOk.WithData(sessionData)

	gctx.JSON(http.StatusOK, result)
}
