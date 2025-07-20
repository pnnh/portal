package articles

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/frontmatter"
	"github.com/sirupsen/logrus"
	"neutron/helpers"
	"portal/models/notes"
	"portal/services/githelper"
)

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
		note := &notes.MTNoteModel{
			Uid:         matter.Uid,
			Title:       noteTitle,
			Body:        string(rest),
			Description: matter.Description,
			Status:      1,
			Cid:         matter.Uid,
			Lang:        "zh",
			Dc:          "hk",
		}

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
		err = notes.PGInsertNote(note)
		if err != nil {
			fmt.Printf("插入文章失败: %v", err)
		}
	}

	return nil
}
