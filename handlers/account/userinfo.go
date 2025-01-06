package account

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"portal/models"
	"portal/neutron/config"
	"portal/neutron/helpers"
)

func UserinfoHandler(gctx *gin.Context) {
	authCookie, err := gctx.Request.Cookie("PT")
	if err != nil && err != http.ErrNoCookie {
		logrus.Errorln("获取cookie失败", err)
		gctx.JSON(http.StatusOK, models.CodeError.WithError(err))
		return
	}

	jwtPublicKey, ok := config.GetConfigurationString("JWT_PUBLIC_KEY")
	if !ok || jwtPublicKey == "" {
		logrus.Errorln("JWT_PUBLIC_KEY 未配置")
	}

	jwtId := ""
	if authCookie != nil && authCookie.Value != "" {
		jwtToken := strings.TrimPrefix(authCookie.Value, "Bearer ")
		parsedClaims, err := helpers.ParseJwtTokenRs256(jwtToken, jwtPublicKey)
		if err != nil && !errors.Is(err, jwt.ErrTokenExpired) {
			gctx.JSON(http.StatusOK, models.CodeError.WithError(err))
			return
		}
		if parsedClaims != nil {
			jwtId = parsedClaims.ID
		}
	}
	if jwtId == "" {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("jwtId为空"))
		return
	}

	accountModel, err := models.GetAccountBySessionId(jwtId)
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
