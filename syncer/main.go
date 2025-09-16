package syncer

import (
	"fmt"
	"sync"
	"time"

	"neutron/config"
	"neutron/services/datastore"
	"neutron/services/filesystem"
	"portal/syncer/articles"

	"github.com/sirupsen/logrus"
)

// 该程序用于定时同步文章和仓库
// 考虑到简化实现，目前仅能够一次性执行，无法作为服务运行
func SyncerMain(configFlag string) {
	logrus.Println("Hello, Syncer!")

	err := config.InitAppConfig(configFlag, "huable", "polaris", config.GetEnvName(), "syncer")
	if err != nil {
		logrus.Fatalln("初始化配置失败", err)
	}

	accountDSN, ok := config.GetConfiguration("app.DATABASE")
	if !ok || accountDSN == nil {
		logrus.Errorln("DATABASE未配置2")
	}

	if err := datastore.Init(accountDSN.(string)); err != nil {
		logrus.Fatalln("datastore: ", err)
	}

	var wg = &sync.WaitGroup{}
	nowTime := time.Now()
	syncno := fmt.Sprintf("SYN%s", nowTime.Format("200601021504"))
	// 仓库同步Worker
	repoWorker, err := articles.NewRepoWorker(wg, syncno)
	if err != nil {
		logrus.Errorln("初始化RepoWorker失败", err)
		return
	}

	wg.Add(1)
	go repoWorker.StartWork()

	blogUrl, ok := config.GetConfigurationString("BLOG_URL")
	if !ok || blogUrl == "" {
		logrus.Fatalln("BLOG_URL 未配置")
	}
	blogDir, err := filesystem.ResolvePath(blogUrl)
	if err != nil {
		logrus.Fatalln("解析路径失败", err)
		return
	}
	sourceUrl, ok := config.GetConfiguration("SOURCE_URL")
	if !ok || sourceUrl == nil {
		logrus.Fatalln("SOURCE_URL 未配置")
	}
	sourceDir, err := filesystem.ResolvePath(sourceUrl.(string))

	if err != nil {
		logrus.Fatalln("解析路径失败", err)
		return
	}

	wg.Add(2)
	go SyncDirectoryForever(repoWorker, sourceDir, wg, syncno)
	go SyncDirectoryForever(repoWorker, blogDir, wg, syncno)

	wg.Wait()
}

func SyncDirectoryForever(repoWorker *articles.RepoWorker, dirPath string, wg *sync.WaitGroup, syncno string) {
	logrus.Println("开始定时同步目录:", dirPath)
	for {
		// 文章同步Worker
		articleWorker, err := articles.NewArticleWorker(repoWorker, dirPath, syncno)
		if err != nil {
			logrus.Errorln("初始化ArticleWorker失败", err)
			return
		}
		articleWorker.StartWork()
		wg.Done()
	}
}
