package config

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	v2 "portal/neutron/config/v2"
)

var staticConfigModel v2.ConfigMap

func InitAppConfig() error {
	logrus.SetReportCaller(false)
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:     false,
		TimestampFormat: time.RFC3339,
	})

	model, err := v2.ParseConfig("config.yaml")
	if err != nil {
		return fmt.Errorf("配置文件解析失败: %w", err)
	}
	staticConfigModel = model

	if Debug() {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
	logrus.Println("日志级别:", logrus.GetLevel())
	return nil
}

func GetConfiguration(key interface{}) (interface{}, bool) {
	if key, ok := key.(string); ok {
		if value, err := staticConfigModel.GetValue(key); err == nil {
			return value, true
		}
	}
	return nil, false
}

func GetConfigurationString(key interface{}) (string, bool) {
	if key, ok := key.(string); ok {
		if value, err := staticConfigModel.GetString(key); err == nil {
			return value, true
		}
	}
	return "", false
}

func MustGetConfigurationString(key interface{}) string {
	value, ok := GetConfigurationString(key)
	if !ok {
		logrus.Fatalf("配置项[%s]不存在", key)
	}
	return value
}

func GetConfigurationInt64(key interface{}) (int64, bool) {
	if key, ok := key.(string); ok {
		if value, err := staticConfigModel.GetInt64(key); err == nil {
			return value, true
		}
	}
	return 0, false
}

func GetConfigOrDefaultInt64(key interface{}, defaultValue int64) int64 {
	value, ok := GetConfigurationInt64(key)
	if !ok {
		return defaultValue
	}
	return value
}

func Debug() bool {
	var mode, ok = GetConfiguration("MODE")
	if ok && mode == "DEBUG" {
		return true
	}
	return false
}
