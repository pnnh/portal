package usercon

import (
	"net/http"

	nemodels "neutron/models"
	"portal/business"
	"portal/models"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// 获取当前登录用户的信息，需要当前登录用户的cookie
func UserinfoHandler(gctx *gin.Context) {
	sessionAccountModel, err := business.FindAccountFromCookie(gctx)
	if err != nil {
		logrus.Warnln("UserinfoHandler", err)
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "查询账号出错b"))
		return
	}
	if sessionAccountModel == nil {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("账号不存在"))
		return
	}
	databaseAccountModel, err := models.GetAccount(sessionAccountModel.Uid)
	if err != nil {
		logrus.Warnln("UserinfoHandler", err)
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "查询账号信息出错"))
		return
	}
	selfAccountModel := &models.SelfAccountModel{
		AccountModel: *databaseAccountModel,
		Username:     sessionAccountModel.Username,
	}

	result := nemodels.NECodeOk.WithData(selfAccountModel)

	gctx.JSON(http.StatusOK, result)
}
