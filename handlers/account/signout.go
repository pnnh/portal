package account

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"portal/business"
	"portal/models"
)

func SignoutHandler(gctx *gin.Context) {
	// 移除cookie
	gctx.SetCookie(business.AuthCookieName, "", -1, "/", "", true, true)

	result := models.CodeOk.WithData(map[string]interface{}{"message": "退出成功"})

	gctx.JSON(http.StatusOK, result)

}
