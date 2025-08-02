package main

import (
	"fmt"
	"net/http"
	"os"
	"portal/business/account"
	"portal/business/comments"
	"portal/business/notes"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"neutron/config"
	"neutron/services/filesystem"
	"portal/business/channels"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"portal/handlers"
)

type IResource interface {
	RegisterRouter(router *gin.Engine, name string)
}

type WebServer struct {
	router    *gin.Engine
	resources map[string]IResource
}

func checkCorsOrigin(origin string) bool {
	if config.Debug() {
		return true // 在调试模式下允许所有来源
	}
	if strings.HasSuffix(origin, "huable.xyz") {
		return true
	}
	return false
}

func NewWebServer() (*WebServer, error) {
	router := gin.Default()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	server := &WebServer{
		router:    router,
		resources: make(map[string]IResource)}

	router.Use(cors.New(cors.Config{
		AllowOriginFunc:  checkCorsOrigin,
		AllowMethods:     []string{"PUT", "PATCH", "POST", "GET"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Portal-Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	return server, nil
}

func (s *WebServer) Init() error {
	indexHandler := handlers.NewIndexHandler()
	s.router.GET("/portal/healthz", indexHandler.Query)

	//authHandler := &handlers.WebauthnHandler{}
	//s.router.POST("/account/signup/webauthn/begin/:username", authHandler.BeginRegistration)
	//s.router.POST("/account/signup/webauthn/finish/:username", authHandler.FinishRegistration)
	//s.router.POST("/account/signin/webauthn/begin/:username", authHandler.BeginLogin)
	//s.router.POST("/account/signin/webauthn/finish/:username", authHandler.FinishLogin)
	//
	//s.router.POST("/account/signup/password/finish", account.PasswordSignupFinishHandler)
	//s.router.POST("/account/signin/password/finish", account.PasswordSigninFinishHandler)

	//s.router.GET("/public/applications", public.PublicApplicationSelectHandler)
	//
	//s.router.GET("/oauth2/auth", func(gctx *gin.Context) {
	//	authorizationserver.AuthEndpointHtml(gctx)
	//})
	//s.router.POST("/oauth2/auth", func(gctx *gin.Context) {
	//	authorizationserver.AuthEndpointJson(gctx)
	//})
	//
	//s.router.POST("/oauth2/token", authorizationserver.TokenEndpoint)
	//s.router.POST("/oauth2/revoke", func(gctx *gin.Context) {
	//	authorizationserver.RevokeEndpoint(gctx)
	//})
	//s.router.POST("/oauth2/introspect", func(gctx *gin.Context) {
	//	authorizationserver.IntrospectionEndpoint(gctx)
	//})
	//s.router.GET("/oauth2/jwks", func(gctx *gin.Context) {
	//	authorizationserver.JwksEndpoint(gctx)
	//})
	//s.router.POST("/oauth2/user", func(gctx *gin.Context) {
	//	authorizationserver.UserEndpoint(gctx)
	//})

	//s.router.GET("/api/go_captcha_data", captcha.GetCaptchaData)
	//s.router.POST("/api/go_captcha_check_data", captcha.CheckCaptcha)

	s.router.Use(gin.Recovery())
	storageUrl, ok := config.GetConfigurationString("STORAGE_URL")
	if !ok || storageUrl == "" {
		return fmt.Errorf("STORAGE_URL 未配置2")
	}
	storagePath, err := filesystem.ResolvePath(storageUrl)
	if err != nil {
		return fmt.Errorf("解析路径失败: %w", err)
	}
	s.router.Static("/portal/storage", storagePath)

	s.router.POST("/portal/comments", comments.CommentInsertHandler)
	s.router.GET("/portal/comments", comments.CommentSelectHandler)
	s.router.GET("/portal/articles", notes.NoteSelectHandler)
	s.router.GET("/portal/console/articles", notes.ConsoleNotesSelectHandler)
	s.router.GET("/portal/:lang/console/articles", notes.ConsoleNotesSelectHandler)
	s.router.POST("/portal/console/articles", notes.NoteConsoleInsertHandler)
	s.router.GET("/portal/articles/:uid", notes.NoteGetHandler)
	s.router.GET("/portal/console/articles/:uid", notes.ConsoleNoteGetHandler)
	s.router.GET("/portal/:lang/console/articles/:uid", notes.ConsoleNoteGetHandler)
	s.router.PUT("/portal/console/articles/:uid", notes.ConsoleNoteUpdateHandler)
	s.router.DELETE("/portal/console/articles/:uid", notes.ConsoleNoteDeleteHandler)
	s.router.GET("/portal/articles/:uid/assets", notes.NoteAssetsSelectHandler)
	s.router.GET("/portal/channels", channels.ChannelSelectHandler)
	s.router.GET("/portal/console/channels", channels.ConsoleChannelSelectHandler)
	s.router.POST("/portal/console/channels", channels.ConsoleChannelInsertHandler)
	s.router.GET("/portal/channels/complete", channels.ChannelCompleteHandler)
	s.router.GET("/portal/channels/:uid", channels.ChannelGetByUidHandler)
	s.router.GET("/portal/:lang/channels/uid/:uid", channels.ChannelGetByUidHandler)
	s.router.GET("/portal/:lang/channels/cid/:cid/:wantLang", channels.ChannelGetByCidHandler)
	s.router.GET("/portal/console/channels/:uid", channels.ConsoleChannelGetHandler)
	s.router.PUT("/portal/console/channels/:uid", channels.ConsoleChannelUpdateHandler)
	s.router.DELETE("/portal/console/channels/:uid", channels.ConsoleChannelDeleteHandler)
	s.router.POST("/portal/articles/:uid/viewer", notes.NoteViewerInsertHandler)
	s.router.POST("/portal/account/signup", account.SignupHandler)
	s.router.POST("/portal/account/signin", account.SigninHandler)
	s.router.POST("/portal/account/signout", account.SignoutHandler)
	s.router.GET("/portal/account/userinfo", account.UserinfoHandler)
	s.router.POST("/portal/account/userinfo/edit", account.UserinfoEditHandler)
	s.router.GET("/portal/account/session", account.SessionQueryHandler)
	s.router.GET("/portal/account/auth/app", account.AppQueryHandler)
	s.router.POST("/portal/account/auth/permit", account.PermitAppLoginHandler)

	s.router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		logrus.Debugln("404路径: " + path)
		c.JSON(404, gin.H{"code": path + "PAGE_NOT_FOUND", "message": "Page not found"})
	})

	return nil
}

func (s *WebServer) Start() error {
	if err := s.Init(); err != nil {
		return fmt.Errorf("初始化出错: %w", err)
	}
	port := os.Getenv("PORT")
	if len(port) < 1 {
		port = "8001"
	}
	var handler http.Handler = s

	serv := &http.Server{
		Addr:           ":" + port,
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	logrus.Println("启动服务: " + port)
	if err := serv.ListenAndServe(); err != nil {
		return fmt.Errorf("服务出错停止: %w", err)
	}
	return nil
}

func (s *WebServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
