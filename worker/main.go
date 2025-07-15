package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"neutron/config"
	"neutron/services/datastore"
	"neutron/services/redisdb"
	"portal/models"
	"portal/models/notes"
)

func main() {
	err := config.InitAppConfig()
	if err != nil {
		logrus.Fatalln("初始化配置失败", err)
	}
	logrus.Println("日志初始化完成")
	redisUrl, ok := config.GetConfigurationString("REDIS_URL")
	if !ok || redisUrl == "" {
		logrus.Fatalln("REDIS_URL未配置")
	}
	redisClient, err := redisdb.ConnectRedis(context.Background(), redisUrl)

	logrus.Println("Redis初始化完成")

	accountDSN, ok := config.GetConfiguration("DATABASE")
	if !ok || accountDSN == nil {
		logrus.Errorln("DATABASE未配置")
	}

	if err := datastore.Init(accountDSN.(string)); err != nil {
		logrus.Fatalln("datastore: ", err)
	}
	logrus.Println("DATABASE初始化完成")

	logrus.Println("开始工作")
	// 生产者：推送一些示例任务
	for {
		contentData, err := redisdb.Consume(context.Background(), redisClient, models.CommentViewersRedisKey)
		if err != nil {
			logrus.Errorln("消费数据失败:", err)
			continue
		}

		commentViewers := make([]*notes.MTViewerModel, 0)
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

func SaveToDatabase(commentViewers []*notes.MTViewerModel) error {
	// 模拟保存到数据库的操作

	opErr, itemErrs := notes.PGInsertViewer(commentViewers...)
	if opErr != nil {
		return fmt.Errorf("PGInsertViewer: %v", opErr)
	}
	for key, item := range itemErrs {
		if !errors.Is(item, notes.ErrViewerLogExists) {
			logrus.Warnln("CommentSelectHandler", key, item)
		}
	}

	return nil
}
