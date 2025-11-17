package account

import (
	"net/http"
	"time"

	nemodels "neutron/models"

	"portal/models"

	"neutron/config"
	"neutron/helpers"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func MailSignupBeginHandler(gctx *gin.Context) {
	username := gctx.PostForm("username")
	nickname := gctx.PostForm("nickname")
	if username == "" {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("account为空"))
		return
	}
	validate := validator.New()
	if err := validate.Var(username, "required,email"); err != nil {
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "account格式错误"))
		return
	}
	accountModel, err := models.GetAccountByUsername(username)
	if err != nil {
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "account不存在"))
		return
	}
	if accountModel == nil {
		accountModel = &models.AccountModel{
			Uid:         helpers.NewPostId(),
			Username:    username,
			Password:    "",
			CreateTime:  time.Now(),
			UpdateTime:  time.Now(),
			Nickname:    nickname,
			EMail:       username,
			Credentials: "",
			Session:     "",
		}
		if err := models.PutAccount(accountModel); err != nil {
			gctx.JSON(http.StatusOK, nemodels.NECodeError.WithError(err))
			return
		}
	} else {
		gctx.JSON(http.StatusOK, nemodels.NECodeAccountExists.WithMessage("账号已存在"))
		return
	}

	mailSender := config.MustGetConfigurationString("MAIL_SENDER")
	if len(mailSender) < 3 {
		gctx.JSON(http.StatusOK, nemodels.NECodeAccountExists.WithMessage("邮箱发送者未配置"))
		return
	}

	session := &models.SessionModel{
		Uid:        helpers.MustUuid(),
		Content:    "",
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
		Username:   accountModel.Uid,
		Type:       "signup",
		Code:       helpers.RandNumberRunes(6),
	}

	if err := models.PutSession(session); err != nil {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithError(err))
		return
	}

	// TODO: 为防止被刷，暂不发送邮件
	// subject := "注册验证码"
	// body := "您的验证码是: " + session.Code
	// err = email.SendMail(mailSender, subject, body, username)
	// if err != nil {
	// 	gctx.JSON(http.StatusOK, nemodels.NECodeError.ToResult())
	// 	return
	// }

	sessionData := map[string]interface{}{
		"session": session.Uid,
	}

	result := nemodels.NECodeOk.WithData(sessionData)

	gctx.JSON(http.StatusOK, result)
}

func MailSignupFinishHandler(gctx *gin.Context) {
	session := gctx.PostForm("session")
	code := gctx.PostForm("code")
	if session == "" || code == "" {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("code或session为空"))
		return
	}
	sessionModel, err := models.GetSessionById(session)
	if err != nil {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithError(err))
		return
	}
	if sessionModel == nil {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("sessionModel不存在"))
		return
	}
	if sessionModel.Code != code {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("验证码错误"))
		return
	}
	sessionData := map[string]interface{}{
		"session": sessionModel.Uid,
	}

	result := nemodels.NECodeOk.WithData(sessionData)

	gctx.JSON(http.StatusOK, result)
}

func MailSigninBeginHandler(gctx *gin.Context) {
	username := gctx.PostForm("username")
	if username == "" {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("account为空"))
		return
	}
	validate := validator.New()
	if err := validate.Var(username, "required,email"); err != nil {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithError(err))
		return
	}
	accountModel, err := models.GetAccountByUsername(username)
	if err != nil {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithError(err))
		return
	}
	if accountModel == nil {
		gctx.JSON(http.StatusOK, nemodels.NECodeAccountNotExists.WithMessage("account不存在"))
		return
	}

	mailSender := config.MustGetConfigurationString("MAIL_SENDER")
	if len(mailSender) < 3 {
		gctx.JSON(http.StatusOK, nemodels.NECodeAccountExists.WithMessage("邮箱发送者未配置"))
		return
	}

	session := &models.SessionModel{
		Uid:        helpers.MustUuid(),
		Content:    "",
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
		Username:   accountModel.Uid,
		Type:       "signin",
		Code:       helpers.RandNumberRunes(6),
	}

	if err := models.PutSession(session); err != nil {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithError(err))
		return
	}

	// TODO: 为防止被刷，暂不发送邮件
	// subject := "登陆验证码"
	// body := "您的验证码是: " + session.Code
	// err = email.SendMail(mailSender, subject, body, username)
	// if err != nil {
	// 	gctx.JSON(http.StatusOK, nemodels.NECodeError.ToResult())
	// 	return
	// }

	sessionData := map[string]interface{}{
		"session": session.Uid,
	}

	result := nemodels.NECodeOk.WithData(sessionData)

	gctx.JSON(http.StatusOK, result)
}
