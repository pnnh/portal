package main

import (
	"time"

	"github.com/sirupsen/logrus"
	"portal/neutron/config"
	"portal/neutron/services/datastore"
	"portal/neutron/services/filesystem"
	"portal/syncer/articles"
	"portal/syncer/images"
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
	repoWorker, err := articles.NewRepoWorker()
	if err != nil {
		logrus.Errorln("初始化RepoWorker失败", err)
		return
	}

	go repoWorker.StartWork()

	blogUrl, ok := config.GetConfigurationString("BLOG_URL")
	if !ok || blogUrl == "" {
		logrus.Fatalln("BLOG_URL 未配置")
	}
	resolvedPath, err := filesystem.ResolvePath(blogUrl)
	if err != nil {
		logrus.Fatalln("解析路径失败", err)
		return
	}
	imagesWorker, err := images.NewSyncImagesWorker(resolvedPath)
	if err != nil {
		logrus.Errorln("初始化ImagesWorker失败", err)
		return
	}
	go imagesWorker.StartWork()

	for {
		// 文章同步Worker
		articleWorker, err := articles.NewArticleWorker(repoWorker)
		if err != nil {
			logrus.Errorln("初始化ArticleWorker失败", err)
			return
		}
		articleWorker.StartWork()
		time.Sleep(time.Minute * 120)
	}
}
