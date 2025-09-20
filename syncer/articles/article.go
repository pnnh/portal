package articles

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"neutron/services/checksum"
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
	syncno     string
}

func NewArticleWorker(repoWorker *RepoWorker, rootPath string, syncno string) (*ArticleWorker, error) {
	return &ArticleWorker{
		repoWorker: repoWorker,
		rootPath:   rootPath,
		syncno:     syncno,
	}, nil
}

func (w *ArticleWorker) StartWork() {
	err := filepath.Walk(w.rootPath, w.visitFile)
	if err != nil {
		logrus.Fatalln("error walking the path", w.rootPath, err)
	}
}

func (w *ArticleWorker) visitFile(path string, info os.FileInfo, err error) error {
	if err != nil {
		return fmt.Errorf("error walking the path %s, %w", path, err)
	}

	fileNmae := strings.ToLower(filepath.Base(path))
	if info.IsDir() && strings.HasPrefix(fileNmae, ".") {
		return filepath.SkipDir
	}
	if IsIgnoredPath(path) {
		return filepath.SkipDir
	}
	if info.IsDir() {
		logrus.Infoln("===visitDir===", path)
	}
	if info.IsDir() || !strings.HasSuffix(fileNmae, ".md") {
		return nil
	}
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
		//logrus.Infoln("这是一个MTNote: ", matter)

		sumValue, err := checksum.CalcSha256(path)
		if err != nil {
			return fmt.Errorf("计算文件校验和失败: %w", err)
		}
		//dbNote, err := notes.PGGetNoteByChecksum(sumValue)
		//if err != nil {
		//	return fmt.Errorf("查询文章失败: %w", err)
		//}
		//if dbNote != nil && dbNote.UpdateTime.After(time.Now().Add(-8*time.Hour)) {
		//	//logrus.Infoln("文章已存在，跳过: ", path)
		//	return nil
		//}
		logrus.Infoln("开始同步文章: ", path)

		baseDir := filepath.Dir(path)
		err = w.repoWorker.AddJob(baseDir)
		if err != nil {
			logrus.Warningln("添加任务失败: %w", err)
		}
		noteTitle := strings.Trim(matter.Title, " \n\r\t ")
		if noteTitle == "" {
			noteTitle = strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))
		}
		// 旧的兼容格式处理
		if noteTitle == "index" {
			parentDir := filepath.Base(baseDir)
			if strings.HasSuffix(parentDir, ".note") {
				noteTitle = strings.TrimSuffix(parentDir, ".note")
			}
		}
		note := &notes.MTNoteTable{}
		note.Uid = matter.Uid
		note.Title = noteTitle
		note.Body = string(rest)
		note.Description = matter.Description
		note.Keywords = matter.Keywords
		note.Channel = sql.NullString{String: matter.Chan, Valid: matter.Chan != ""}
		note.UpdateTime = time.Now()
		note.Status = 1 // 已发布
		note.Header = "MTNote"
		note.Lang = "zh"
		note.Name = strcase.ToKebab(noteTitle)
		note.Owner = SyncerArticleOwner
		note.Checksum = sql.NullString{String: sumValue, Valid: true}
		note.Syncno = sql.NullString{String: w.syncno, Valid: true}

		gitInfo, err := githelper.GitInfoGet(baseDir)
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
			relativePath := strings.TrimPrefix(path, gitInfo.RootPath)
			note.RelativePath = sql.NullString{String: relativePath, Valid: true}
			//note.RepoId = sql.NullString{String: gitInfo.RepoId, Valid: true}
			note.RepoFirstCommit = sql.NullString{String: gitInfo.FirstCommitId, Valid: true}
		}
		err = notes.PGConsoleInsertNote(note)
		if err != nil {
			logrus.Errorf("插入文章失败: %v", err)
		}
	}

	return nil
}
