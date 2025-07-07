package account

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"portal/models"
)

func SessionQueryHandler(gctx *gin.Context) {
	link := gctx.Query("link")
	session := gctx.Query("session")
	app := gctx.Query("app")

	var sessionAccountModel *models.SessionModel
	var err error
	if session != "" {
		sessionAccountModel, err = models.GetSessionById(session)
		if err != nil {
			logrus.Warnln("UserinfoHandler2", err)
			gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "查询账号出错b"))
			return
		}
	} else if link != "" && app != "" {
		sessionAccountModel, err = models.GetSessionByLink(app, link)
		if err != nil {
			logrus.Warnln("UserinfoHandler3", err)
			gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "查询账号出错e"))
			return
		}
	} else {
		gctx.JSON(http.StatusBadRequest, models.CodeError.WithMessage("parameters invalid"))
		return
	}
	if sessionAccountModel == nil {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("账号不存在"))
		return
	}
	databaseAccountModel, err := models.GetAccount(sessionAccountModel.Account)
	if err != nil || databaseAccountModel == nil {
		logrus.Warnln("UserinfoHandler", err)
		gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "查询账号信息出错"))
		return
	}
	selfAccountModel := &models.SelfAccountModel{
		AccountModel: *databaseAccountModel,
		Username:     sessionAccountModel.Username,
	}

	result := models.CodeOk.WithData(selfAccountModel)

	gctx.JSON(http.StatusOK, result)
}
