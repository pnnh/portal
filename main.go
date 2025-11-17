package main

import (
	"flag"
	"os"
	"strings"

	"portal/syncer"
	"portal/worker"

	"neutron/config"
	"neutron/services/datastore"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var (
	configFlag  string
	svcroleFlag string
)

func init() {
	flag.StringVar(&configFlag, "config", "file://config.yaml", "config file path")
	flag.StringVar(&svcroleFlag, "svcrole", "portal", "service role, default is portal")
}

func main() {
	flag.Parse()
	logrus.Println("config:", configFlag)
	if strings.HasPrefix(configFlag, "env://") {
		envName := configFlag[len("env://"):]
		configFlag = os.Getenv(envName)
	}
	if configFlag == "" {
		logrus.Fatalln("please set config")
		return
	}

	if config.Debug() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	switch svcroleFlag {
	case "worker":
		logrus.Println("portal worker mode")
		worker.WorkerMain(configFlag)
	case "syncer":
		logrus.Println("portal syncer mode")
		syncer.SyncerMain(configFlag)
	default:
		logrus.Println("portal main mode")
		PortalMain()
	}

}

func PortalMain() {

	err := config.InitAppConfig(configFlag, "huable", "polaris", config.GetEnvName(), "portal")
	if err != nil {
		logrus.Fatalln("初始化配置失败1", err)
	}

	accountDSN, ok := config.GetConfiguration("app.DATABASE")
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
