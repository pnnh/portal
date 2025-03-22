package main

import (
	"time"

	"github.com/sirupsen/logrus"
	"portal/quark/neutron/config"
	"portal/quark/neutron/services/datastore"
)

func main() {
	logrus.Println("Hello, Syncer!")

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

	// 仓库同步Worker
	repoWorker, err := NewRepoWorker()
	if err != nil {
		logrus.Errorln("初始化RepoWorker失败", err)
		return
	}

	go repoWorker.StartWork()

	for {
		// 文章同步Worker
		articleWorker, err := NewArticleWorker(repoWorker)
		if err != nil {
			logrus.Errorln("初始化ArticleWorker失败", err)
			return
		}
		articleWorker.StartWork()
		time.Sleep(time.Minute * 5)

	}
}
