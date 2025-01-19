package main

import (
	"database/sql"
	"fmt"
	"github.com/adrg/frontmatter"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"portal/models/notes"
	"portal/neutron/config"
	"portal/neutron/helpers"
	"portal/neutron/services/datastore"
	"portal/neutron/services/filesystem"
	"strings"
)

func visit(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if info.IsDir() {
		readmeFilePath := filepath.Join(path, "README.md")
		if _, err := os.Stat(readmeFilePath); os.IsNotExist(err) {
			return nil
		} else {
			noteText, err := os.ReadFile(readmeFilePath)
			if err != nil {
				return fmt.Errorf("读取文件失败: %w", err)
			}
			matter := &notes.MTNoteMatter{}
			rest, err := frontmatter.Parse(strings.NewReader(string(noteText)), matter)
			if err != nil {
				return fmt.Errorf("解析文章元数据失败: %w", err)
			}
			if matter.Cls == "MTNote" && helpers.IsUuid(matter.Uid) {
				fmt.Printf("%+v", matter)
				fmt.Println("这是一个MTNote")

				note := &notes.MTNoteModel{
					Uid:         matter.Uid,
					Title:       matter.Title,
					Body:        string(rest),
					Description: matter.Description,
					Version:     sql.NullString{String: "", Valid: true},
					Build:       sql.NullString{String: "", Valid: true},
					Url:         sql.NullString{String: "", Valid: true},
				}
				err = notes.PGInsertNote(note)
				if err != nil {
					fmt.Printf("插入文章失败: %w", err)
				}
			}
		}
	}
	return nil
}

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

	sourceUrl, ok := config.GetConfiguration("SOURCE_URL")
	if !ok || sourceUrl == nil {
		logrus.Errorln("SOURCE_URL 未配置")
	}
	resolvedPath, err := filesystem.ResolvePath(sourceUrl.(string))
	if err != nil {
		logrus.Errorln("解析路径失败", err)
		return
	}

	err = filepath.Walk(resolvedPath, visit)
	if err != nil {
		fmt.Printf("error walking the path %v: %v\n", resolvedPath, err)
	}
}
