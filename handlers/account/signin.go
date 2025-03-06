package account

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"portal/business"
	"portal/business/cloudflare"
	"portal/models"
	"portal/neutron/config"
	"portal/neutron/helpers"
)

type SigninRequest struct {
	cloudflare.TurnstileModel
	Username    string `json:"username"`    // 账号
	Password    string `json:"password"`    // 密码
	Fingerprint string `json:"fingerprint"` // 指纹
}

func SigninHandler(gctx *gin.Context) {
	request := &SigninRequest{}
	if err := gctx.ShouldBindJSON(request); err != nil {
		gctx.JSON(http.StatusOK, models.CodeError.WithError(err))
		return
	}
	if request.Username == "" || request.Password == "" {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("账号或密码为空"))
		return
	}

	ipAddr := helpers.GetIpAddress(gctx)
	verifyOk, err := cloudflare.VerifyTurnstileToken(request.TurnstileModel.TurnstileToken, ipAddr)
	if err != nil || !verifyOk {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("Signin验证出错"))
		return
	}

	accountModel, err := models.GetAccountByUsername(request.Username)
	if err != nil {
		logrus.Println("GetAccountByUsername", err)
		gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "查询账号出错a"))
		return
	}
	if accountModel == nil {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("账号不存在"))
		return
	}
	if !helpers.CheckPasswordHash(request.Password, accountModel.Password) {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("密码错误"))
		return
	}

	sessionModel := &models.SessionModel{
		Uid:          helpers.MustUuid(),
		Content:      "",
		CreateTime:   time.Now(),
		UpdateTime:   time.Now(),
		Username:     request.Username,
		Type:         "signin",
		Code:         "",
		ClientId:     "",
		ResponseType: "",
		RedirectUri:  "",
		Scope:        "",
		State:        "",
		Nonce:        "",
		IdToken:      "",
		AccessToken:  "",
		JwtId:        "",
		Account:      accountModel.Uid,
	}
	err = models.PutSession(sessionModel)
	if err != nil {
		logrus.Println("PutSession", err)
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("更新会话错误"))
		return
	}

	jwtPrivateKey, ok := config.GetConfigurationString("JWT_PRIVATE_KEY")
	if !ok || jwtPrivateKey == "" {
		logrus.Errorln("JWT_PRIVATE_KEY 未配置")
	}

	jwtToken, err := helpers.GenerateJwtTokenRs256(sessionModel.Username, jwtPrivateKey, sessionModel.Uid)
	if (jwtToken == "") || (err != nil) {
		logrus.Println("GenerateJwtTokenRs256", err)
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("生成jwtToken错误"))
		return
	}

	// 登录成功后设置cookie
	gctx.SetCookie(business.AuthCookieName, jwtToken, 3600*48, "/", "", true, true)

	result := models.CodeOk.WithData(map[string]any{
		"changes": 1,
		"uid":     sessionModel.Uid,
	})

	gctx.JSON(http.StatusOK, result)
}
