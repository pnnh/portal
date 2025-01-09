package account

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"portal/business"
	"portal/models"
)

func UserinfoHandler(gctx *gin.Context) {
	accountModel, err := business.FindUserFromCookie(gctx)
	if err != nil {
		logrus.Println("GetAccountBySessionId", err)
		gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "查询账号出错"))
		return
	}
	if accountModel == nil {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("账号不存在"))
		return
	}

	result := models.CodeOk.WithData(accountModel)

	gctx.JSON(http.StatusOK, result)
}
