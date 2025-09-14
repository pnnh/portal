package articles

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"portal/business/notes"

	"github.com/iancoleman/strcase"

	"neutron/helpers"
	"portal/services/githelper"

	"github.com/adrg/frontmatter"
	"github.com/sirupsen/logrus"
)

// 一个固定的用户ID，通过Syncer服务同步的文章属于这个用户
const SyncerArticleOwner = "01990e6a-2689-731b-a5a2-b46117e22040"

type ArticleWorker struct {
	repoWorker *RepoWorker
	rootPath   string
}

func NewArticleWorker(repoWorker *RepoWorker, rootPath string) (*ArticleWorker, error) {
	return &ArticleWorker{
		repoWorker: repoWorker,
		rootPath:   rootPath,
	}, nil
}

func (w *ArticleWorker) StartWork() {

	err := filepath.Walk(w.rootPath, w.visitFile)
	if err != nil {
		logrus.Fatalln("error walking the path", w.rootPath, err)
	}
}

func (w *ArticleWorker) visitFile(path string, info os.FileInfo, err error) error {
	logrus.Infoln("====", path)
	if err != nil {
		return err
	}

	fileNmae := strings.ToLower(filepath.Base(path))
	if info.IsDir() && strings.HasPrefix(fileNmae, ".") {
		return filepath.SkipDir
	}
	if info.IsDir() || !strings.HasSuffix(fileNmae, ".md") {
		return nil
	}

	logrus.Infoln("++++", path)

	noteText, err := os.ReadFile(path)
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

		//err = w.repoWorker.AddJob(path)
		//if err != nil {
		//	logrus.Warningln("添加任务失败: %w", err)
		//}
		noteTitle := strings.Trim(matter.Title, " \n\r\t ")
		if noteTitle == "" {
			noteTitle = strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))
		}
		// 旧的兼容格式处理
		if noteTitle == "index" {
			parentDir := filepath.Base(filepath.Dir(path))
			if strings.HasSuffix(parentDir, ".note") {
				noteTitle = strings.TrimSuffix(parentDir, ".note")
			}
		}
		note := &notes.MTNoteTable{}
		note.Uid = matter.Uid
		note.Title = noteTitle
		note.Body = string(rest)
		note.Description = matter.Description
		note.Status = 1 // 已发布
		note.Header = "MTNote"
		note.Lang = "zh"
		note.Name = strcase.ToKebab(noteTitle)
		note.Owner = SyncerArticleOwner

		gitInfo, err := githelper.GitInfoGet(path)
		if err != nil {
			logrus.Warningln("获取git信息失败: %w", err)
		}
		if gitInfo != nil {
			note.Version = sql.NullString{String: gitInfo.CommitId, Valid: true}
			note.Build = sql.NullString{String: "", Valid: true}
			note.Url = sql.NullString{String: gitInfo.RemoteUrl, Valid: true}
			note.Branch = sql.NullString{String: gitInfo.Branch, Valid: true}
			note.Commit = sql.NullString{String: gitInfo.CommitId, Valid: true}
			note.CommitTime = sql.NullTime{Time: gitInfo.CommitTime, Valid: true}
			repoPath := strings.TrimPrefix(path, gitInfo.RootPath)
			note.RepoPath = sql.NullString{String: repoPath, Valid: true}
			note.RepoId = sql.NullString{String: gitInfo.RepoId, Valid: true}
		}
		err = notes.PGConsoleInsertNote(note)
		if err != nil {
			fmt.Printf("插入文章失败: %v", err)
		}
	} else {
		logrus.Infoln("跳过非MTNote文章", path)
	}

	return nil
}
