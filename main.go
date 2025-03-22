package main

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"portal/quark/neutron/config"
	"portal/quark/neutron/services/datastore"
)

func main() {
	if config.Debug() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	err := config.InitAppConfig()
	if err != nil {
		logrus.Fatalln("初始化配置失败", err)
	}

	accountDSN, ok := config.GetConfiguration("DATABASE")
	if !ok || accountDSN == nil {
		logrus.Errorln("DATABASE未配置")
	}

	if err := datastore.Init(accountDSN.(string)); err != nil {
		logrus.Fatalln("datastore: ", err)
	}

	webServer, err := NewWebServer()
	if err != nil {
		logrus.Fatalln("创建web server出错", err)
	}

	if err := webServer.Start(); err != nil {
		logrus.Fatalln("应用程序执行出错: %w", err)
	}
}
