package account

import (
	"encoding/base64"
	"net/http"
	"time"

	nemodels "neutron/models"

	"neutron/config"
	"portal/handlers/auth/authorizationserver"

	"github.com/sirupsen/logrus"

	"portal/models"

	"neutron/helpers"

	"github.com/gin-gonic/gin"
)

//func PasswordSignupBeginHandler(gctx *gin.Context) {
// username := gctx.PostForm("username")
// nickname := gctx.PostForm("nickname")
// if username == "" {
// 	gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("account为空"))
// 	return
// }
// accountModel, err := models.GetAccountByUsername(username)
// if err != nil {
// 	gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "account不存在"))
// 	return
// }
// if accountModel == nil {
// 	accountModel = &models.AccountModel{
// 		Uid:          helpers.NewPostId(),
// 		Username:    username,
// 		Password:    "",
// 		CreateTime:  time.Now(),
// 		UpdateTime:  time.Now(),
// 		Nickname:    nickname,
// 		EMail:        username,
// 		Credentials: "",
// 		Session:     "",
// 	}
// 	if err := models.PutAccount(accountModel); err != nil {
// 		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithError(err))
// 		return
// 	}
// } else {
// 	gctx.JSON(http.StatusOK, models.CodeAccountExists.WithMessage("账号已存在"))
// 	return
// }

// session := &models.SessionModel{
// 	Uid:         helpers.MustUuid(),
// 	Content:    "",
// 	CreateTime: time.Now(),
// 	UpdateTime: time.Now(),
// 	Username:   accountModel.Uid,
// 	Type:       "signup_password",
// 	Code:       helpers.RandNumberRunes(6),
// }

// if err := models.PutSession(session); err != nil {
// 	gctx.JSON(http.StatusOK, nemodels.NECodeError.WithError(err))
// 	return
// }

// sessionData := map[string]interface{}{
// 	"session": session.Uid,
// }

// result := nemodels.NECodeOk.WithData(sessionData)

// gctx.JSON(http.StatusOK, result)
//}

func PasswordSignupFinishHandler(gctx *gin.Context) {
	username := gctx.PostForm("username")
	password := gctx.PostForm("password")
	source, _ := gctx.GetQuery("source")
	if username == "" || password == "" || source == "" {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("code或session为空"))
		return
	}
	// sessionModel, err := models.GetSessionById(session)
	// if err != nil {
	// 	gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "GetSessionById error"))
	// 	return
	// }
	// if sessionModel == nil {
	// 	gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("sessionModel不存在"))
	// 	return
	// }
	// if sessionModel.Type != "signup_password" {
	// 	gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("sessionModel类型不对"))
	// 	return
	// }
	accountModel, err := models.GetAccountByUsername(username)
	if err != nil {
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "GetAccount error"))
		return
	}
	if accountModel != nil {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("账号已存在"))
		return
	}

	encrypted, err := helpers.HashPassword(password)
	if err != nil {
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "HashPassword error"))
		return
	}

	accountModel = &models.AccountModel{
		Uid:         helpers.NewPostId(),
		Username:    username,
		Password:    encrypted,
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
		Nickname:    username,
		EMail:       username,
		Credentials: "",
		Session:     "",
	}
	if err := models.PutAccount(accountModel); err != nil {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithError(err))
		return
	}

	// if err := models.UpdateAccountPassword(username, encrypted); err != nil {
	// 	gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "UpdateAccountPassword error"))
	// 	return
	// }

	sourceData, err := base64.URLEncoding.DecodeString(source)
	if err != nil {
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "source解析失败"))
		return
	}
	sourceUrl := string(sourceData)
	if len(sourceUrl) < 1 {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("sourceUrl为空"))
		return
	}
	gctx.Redirect(http.StatusFound, sourceUrl)
}

func PasswordSigninFinishHandler(gctx *gin.Context) {
	source, _ := gctx.GetQuery("source")
	//session := gctx.PostForm("session")
	username := gctx.PostForm("username")
	password := gctx.PostForm("password")
	if username == "" || password == "" || source == "" {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("code或session为空2"))
		return
	}

	captchaKey := gctx.PostForm("captcha_key")
	if captchaKey == "" {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("captcha_key为空"))
		return
	}

	captchaModel, err := models.FindCaptcha(captchaKey)
	if captchaModel == nil || err != nil || captchaModel.Checked != 1 ||
		time.Now().Sub(captchaModel.CreateTime).Minutes() > 5 {
		logrus.Errorln("验证码错误", captchaKey, err)
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "验证码错误"))
		return
	}

	account, err := models.GetAccountByUsername(username)
	if err != nil {
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "GetAccount error"))
		return
	}
	if account == nil {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("account不存在"))
		return
	}

	ok := helpers.CheckPasswordHash(password, account.Password)

	if !ok {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("密码错误"))
		return
	}

	jwkString, ok := config.GetConfigurationString("OAUTH2_JWK")
	if !ok {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("OAUTH2_JWK is not set"))
		return
	}
	jwkModel, err := helpers.GetJwkModel(jwkString)
	if err != nil || jwkModel == nil {
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "GetJwkModel error"))
		return
	}

	session := &models.SessionModel{
		Uid:          helpers.NewPostId(),
		Content:      "",
		CreateTime:   time.Now(),
		UpdateTime:   time.Now(),
		Username:     username,
		Type:         "password",
		Code:         "",
		ClientId:     "",
		ResponseType: "",
		RedirectUri:  "",
		Scope:        "",
		State:        "",
		Nonce:        "",
		IdToken:      "",
		AccessToken:  "",
		JwtId:        jwkModel.Kid,
	}
	err = models.PutSession(session)
	if err != nil {
		logrus.Printf("Error occurred in NewAccessResponse2222: %+v", err)
		return
	}

	issuer := config.MustGetConfigurationString("app.PUBLIC_PORTAL_URL")
	jwtToken, err := helpers.GenerateJwtTokenRs256(account.Username,
		authorizationserver.PrivateKeyString,
		session.JwtId, issuer)
	if (jwtToken == "") || (err != nil) {
		models.ResponseMessageError(gctx, "参数有误316", err)
		return
	}

	// 登录成功后设置cookie
	gctx.SetCookie("Portal-Authorization", jwtToken, 3600*72, "/", "", true, true)

	// sessionData := map[string]interface{}{
	// 	"authorization": jwtToken,
	// }

	sourceData, err := base64.URLEncoding.DecodeString(source)
	if err != nil {
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "source解析失败"))
		return
	}
	sourceUrl := string(sourceData)
	if len(sourceUrl) < 1 {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("sourceUrl为空"))
		return
	}
	gctx.Redirect(http.StatusFound, sourceUrl)

	// result := nemodels.NECodeOk.WithData(sessionData)

	// gctx.JSON(http.StatusOK, result)
}
