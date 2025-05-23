package articles

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/frontmatter"
	"github.com/sirupsen/logrus"
	"portal/models/notes"
	"portal/neutron/config"
	"portal/neutron/helpers"
	"portal/neutron/services/filesystem"
	"portal/services/githelper"
)

type ArticleWorker struct {
	repoWorker *RepoWorker
}

func NewArticleWorker(repoWorker *RepoWorker) (*ArticleWorker, error) {
	return &ArticleWorker{
		repoWorker: repoWorker,
	}, nil
}

func (w *ArticleWorker) StartWork() {

	sourceUrl, ok := config.GetConfiguration("SOURCE_URL")
	if !ok || sourceUrl == nil {
		logrus.Fatalln("SOURCE_URL 未配置")
	}
	resolvedPath, err := filesystem.ResolvePath(sourceUrl.(string))
	if err != nil {
		logrus.Fatalln("解析路径失败", err)
		return
	}
	err = filepath.Walk(resolvedPath, w.visitFile)
	if err != nil {
		logrus.Fatalln("error walking the path %v: %v\n", resolvedPath, err)
	}
}

func (w *ArticleWorker) visitFile(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return nil
	}
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

			err = w.repoWorker.AddJob(path)
			if err != nil {
				logrus.Warningln("添加任务失败: %w", err)
			}

			gitInfo, err := githelper.GitInfoGet(path)
			if err != nil {
				fmt.Printf("获取git信息失败: %w", err)
			}

			note := &notes.MTNoteModel{
				Uid:         matter.Uid,
				Title:       matter.Title,
				Body:        string(rest),
				Description: matter.Description,
				Status:      1,
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
			err = notes.PGInsertNote(note)
			if err != nil {
				fmt.Printf("插入文章失败: %w", err)
			}
		}
	}

	return nil
}
