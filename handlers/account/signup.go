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

type SignupRequest struct {
	cloudflare.TurnstileModel
	Username       string `json:"username"`         // 账号
	Password       string `json:"password"`         // 密码
	ConfimPassword string `json:"confirm_password"` // 确认密码
	Nickname       string `json:"nickname"`         // 昵称
	EMail          string `json:"email"`            // 邮箱
	Fingerprint    string `json:"fingerprint"`      // 指纹
}

func SignupHandler(gctx *gin.Context) {
	request := &SignupRequest{}
	if err := gctx.ShouldBindJSON(request); err != nil {
		gctx.JSON(http.StatusOK, models.CodeError.WithError(err))
		return
	}
	if request.Username == "" || request.Password == "" {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("账号或密码为空"))
		return
	}
	if request.Password != request.ConfimPassword {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("两次密码不一致"))
		return
	}

	ipAddr := helpers.GetIpAddress(gctx)
	verifyOk, err := cloudflare.VerifyTurnstileToken(request.TurnstileModel.TurnstileToken, ipAddr)
	if err != nil || !verifyOk {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("验证出错"))
		return
	}

	isExist, err := models.CheckAccountExists(request.Username)
	if err != nil {
		logrus.Println("CheckAccountExists", err)
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("查询账户出错"))
		return
	}
	if isExist {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("账户已存在"))
		return
	}

	hashPassword, err := helpers.HashPassword(request.Password)
	if err != nil {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("密码加密出错"))
		return
	}
	accountModel := &models.AccountModel{
		Urn:         helpers.MustUuid(),
		Username:    request.Username,
		Password:    hashPassword,
		Photo:       "",
		CreateTime:  time.Now().UTC(),
		UpdateTime:  time.Now().UTC(),
		Nickname:    request.Nickname,
		EMail:       request.EMail,
		Credentials: "",
		Session:     "",
		Description: "",
		Status:      0,
		Website:     "",
		Fingerprint: request.Fingerprint,
	}
	err = models.EnsureAccount(accountModel)
	if err != nil || accountModel == nil {
		gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "更新用户账户出错"))
		return
	}

	sessionModel := &models.SessionModel{
		Urn:          helpers.MustUuid(),
		Content:      "",
		CreateTime:   time.Now(),
		UpdateTime:   time.Now(),
		Username:     request.Username,
		Type:         "signup",
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
		Account:      accountModel.Urn,
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

	jwtToken, err := helpers.GenerateJwtTokenRs256(sessionModel.Username, jwtPrivateKey, sessionModel.Urn)
	if (jwtToken == "") || (err != nil) {
		logrus.Println("GenerateJwtTokenRs256", err)
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("生成jwtToken错误"))
		return
	}

	// 登录成功后设置cookie
	gctx.SetCookie(business.AuthCookieName, jwtToken, 3600*48, "/", "", true, true)

	result := models.CodeOk.WithData(map[string]any{
		"changes": 1,
		"urn":     accountModel.Urn,
	})

	gctx.JSON(http.StatusOK, result)
}
