package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"portal/business/comments"
	"portal/business/viewers"

	"neutron/config"
	"neutron/services/datastore"
	"neutron/services/redisdb"

	"github.com/sirupsen/logrus"
)

func WorkerMain(configFlag string) {

	err := config.InitAppConfig(configFlag, "huable", "polaris", config.GetEnvName(), "worker")
	if err != nil {
		logrus.Fatalln("初始化配置失败3", err)
	}

	redisUrl, ok := config.GetConfigurationString("REDIS_URL")
	if !ok || redisUrl == "" {
		logrus.Fatalln("REDIS_URL not found in configuration")
	}
	redisClient, err := redisdb.ConnectRedis(context.Background(), redisUrl)
	if err != nil {
		logrus.Fatalln("redisdb.ConnectRedis error:", err)
	}

	logrus.Println("Redis初始化完成")

	accountDSN, ok := config.GetConfiguration("app.DATABASE")
	if !ok || accountDSN == nil {
		logrus.Errorln("DATABASE未配置3")
	}

	if err := datastore.Init(accountDSN.(string)); err != nil {
		logrus.Fatalln("datastore: ", err)
	}
	logrus.Println("DATABASE初始化完成")

	logrus.Println("Starting comment viewer worker...")
	for {
		contentData, err := redisdb.Consume(context.Background(), redisClient, comments.CommentViewersRedisKey)
		if err != nil {
			logrus.Errorln("消费数据失败:", err)
			continue
		}

		commentViewers := make([]*viewers.MTViewerTable, 0)
		if err := json.Unmarshal(contentData, &commentViewers); err != nil {
			logrus.Errorln("commentViewers Unmarshal error:", err)
			continue
		}
		if len(commentViewers) == 0 {
			continue
		}
		if err := SaveToDatabase(commentViewers); err != nil {
			logrus.Errorln("保存到数据库失败:", err, string(contentData))
			continue
		}

		// 模拟生产间隔
		time.Sleep(1 * time.Second)
	}
}

func SaveToDatabase(commentViewers []*viewers.MTViewerTable) error {
	// 模拟保存到数据库的操作

	opErr, itemErrs := viewers.PGInsertViewer(commentViewers...)
	if opErr != nil {
		return fmt.Errorf("PGInsertViewer: %v", opErr)
	}
	for key, item := range itemErrs {
		if !errors.Is(item, viewers.ErrViewerLogExists) {
			logrus.Warnln("CommentSelectHandler", key, item)
		}
	}

	return nil
}
