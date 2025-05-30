package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"portal/models/images"
	"portal/models/notes"
	"portal/neutron/config"
	"portal/neutron/services/filesystem"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"portal/handlers"
	"portal/handlers/account"
	"portal/handlers/comments"
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

func suzakuProxy(c *gin.Context) {

	defer func() {
		if p := recover(); p != nil {
			logrus.Errorln("suzakuProxy panic: ", p)
		}
	}()
	remote, err := url.Parse("http://127.0.0.1:7102")
	if err != nil {
		logrus.Fatalln("解析远程地址失败: ", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.Director = func(req *http.Request) {
		req.Header = c.Request.Header
		req.Host = remote.Host
		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host
		req.URL.Path = "/suzaku" + c.Param("proxyPath")
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}

func lightningProxy(c *gin.Context) {

	defer func() {
		if p := recover(); p != nil {
			logrus.Errorln("lightningProxy panic: ", p)
		}
	}()
	remote, err := url.Parse("http://localhost:5173")
	if err != nil {
		logrus.Fatalln("解析远程地址失败2: ", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.Director = func(req *http.Request) {
		req.Header = c.Request.Header
		req.Host = remote.Host
		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host
		req.URL.Path = "/lightning" + c.Param("proxyPath")
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}

func polarisProxy(c *gin.Context) {

	defer func() {
		if p := recover(); p != nil {
			logrus.Errorln("polarisProxy panic: ", p)
		}
	}()
	remote, err := url.Parse("http://127.0.0.1:7100")
	if err != nil {
		logrus.Fatalln("解析远程地址失败3: ", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.Director = func(req *http.Request) {
		req.Header = c.Request.Header
		req.Host = remote.Host
		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host
		req.URL.Path = c.Request.URL.Path
	}
	proxy.ServeHTTP(c.Writer, c.Request)
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
	s.router.GET("/portal/articles/:uid", notes.NoteGetHandler)
	s.router.GET("/portal/articles/:uid/assets", notes.NoteAssetsSelectHandler)
	s.router.GET("/portal/images", images.ImageSelectHandler)
	s.router.GET("/portal/images/:uid", images.ImageGetHandler)
	s.router.POST("/portal/articles/:uid/viewer", notes.NoteViewerInsertHandler)
	s.router.POST("/portal/account/signup", account.SignupHandler)
	s.router.POST("/portal/account/signin", account.SigninHandler)
	s.router.POST("/portal/account/signout", account.SignoutHandler)
	s.router.GET("/portal/account/userinfo", account.UserinfoHandler)
	s.router.POST("/portal/account/userinfo/edit", account.UserinfoEditHandler)

	if config.Debug() {
		//s.router.Any("/suzaku/*proxyPath", suzakuProxy)
		//s.router.Any("/lightning/*proxyPath", lightningProxy)
		//s.router.NoRoute(polarisProxy)
	} else {
		s.router.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path
			logrus.Debugln("404路径: " + path)
			c.JSON(404, gin.H{"code": path + "PAGE_NOT_FOUND", "message": "Page not found"})
		})
	}

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
