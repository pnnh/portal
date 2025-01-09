package business

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"portal/models"
	"portal/neutron/config"
	"portal/neutron/helpers"
)

const AuthCookieName = "PT"

func FindUserFromCookie(gctx *gin.Context) (*models.AccountModel, error) {
	authCookie, err := gctx.Request.Cookie(AuthCookieName)
	if err != nil && !errors.Is(err, http.ErrNoCookie) {
		return nil, fmt.Errorf("获取cookie失败: %s", err)
	}

	jwtPublicKey, ok := config.GetConfigurationString("JWT_PUBLIC_KEY")
	if !ok || jwtPublicKey == "" {
		return nil, fmt.Errorf("JWT_PUBLIC_KEY 未配置")
	}

	jwtId := ""
	if authCookie != nil && authCookie.Value != "" {
		jwtToken := strings.TrimPrefix(authCookie.Value, "Bearer ")
		parsedClaims, err := helpers.ParseJwtTokenRs256(jwtToken, jwtPublicKey)
		if err != nil && !errors.Is(err, jwt.ErrTokenExpired) {
			return nil, fmt.Errorf("解析jwtToken失败: %s", err)
		}
		if parsedClaims != nil {
			jwtId = parsedClaims.ID
		}
	}
	if jwtId == "" {
		return nil, fmt.Errorf("jwtId为空")
	}

	accountModel, err := models.GetAccountBySessionId(jwtId)
	if err != nil {
		return nil, fmt.Errorf("查询账号出错: %s", err)
	}

	return accountModel, nil
}
