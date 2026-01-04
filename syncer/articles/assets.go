package articles

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pnnh/neutron/config"
	"github.com/pnnh/neutron/helpers"
	"github.com/pnnh/neutron/services/checksum"
	"github.com/pnnh/neutron/services/filesystem"
	"portal/models/repo"
	"portal/services/githelper"

	"github.com/sirupsen/logrus"
)

type SyncJob struct {
	RepoRootPath string
	GitInfo      *githelper.GitInfo
	filePorter   *filesystem.FilePorter
	ignoreHelper *githelper.GitIgnoreHelper
	syncno       string
}

func NewSyncJob(repoPath string, filePorter *filesystem.FilePorter, syncno string) (*SyncJob, error) {
	gitInfo, err := githelper.GitInfoGet(repoPath)
	if err != nil {
		logrus.Println("获取git信息失败: ", repoPath, err)
		return nil, fmt.Errorf("获取git信息失败: %w", err)
	}
	ignoreHelper := githelper.NewGitIgnoreHelper(repoPath)
	return &SyncJob{
		RepoRootPath: repoPath,
		GitInfo:      gitInfo,
		ignoreHelper: ignoreHelper,
		filePorter:   filePorter,
		syncno:       syncno,
	}, nil
}

func (j *SyncJob) Sync(wg *sync.WaitGroup) {
	defer func() {
		logrus.Infoln("同步完成: ", j.RepoRootPath)
		wg.Done()
	}()
	logrus.Infoln("开始同步仓库: ", j.RepoRootPath)
	// 如果当前commit不是clean状态，则不需要同步
	//if !j.GitInfo.IsClean {
	//	logrus.Println("工作区不干净跳过同步: ", j.RepoRootPath)
	//	return
	//}
	//repoSyncInfo, err := repo.PGGetRepoSyncInfo(j.GitInfo.FirstCommitId, j.GitInfo.Branch)
	//if err != nil {
	//	logrus.Fatalln("获取repo sync info失败: ", j.RepoRootPath, err)
	//	return
	//}
	//if repoSyncInfo != nil && repoSyncInfo.LastCommitId != "" {
	//	isAncestor, err := githelper.GitCommitIsAncestor(j.RepoRootPath, repoSyncInfo.LastCommitId, j.GitInfo.CommitId)
	//	if err != nil {
	//		logrus.Println("比较commit失败: ", j.RepoRootPath, err)
	//		return
	//	}
	//	// 如果当前commit在上次commit之前，则不需要同步
	//	if !isAncestor {
	//		logrus.Println("状态已最新无需同步: ", j.RepoRootPath)
	//		return
	//	}
	//}
	_, err := j.CopyFiles()
	if err != nil {
		logrus.Fatalln("CopyFiles: ", err)
		return
	}
	//repoSyncInfo := &repo.MTRepoSyncModel{
	//	Uid:          helpers.MustUuid(),
	//	LastCommitId: j.GitInfo.CommitId,
	//	Branch:       j.GitInfo.Branch,
	//	//RepoId:       j.GitInfo.RepoId,
	//	FirstCommitId: j.GitInfo.FirstCommitId,
	//}
	//err = repo.PGInsertOrUpdateRepoSyncInfo(repoSyncInfo)
	//if err != nil {
	//	logrus.Fatalln("PGInsertOrUpdateRepoSyncInfo: ", err)
	//	return
	//}
}

func (j *SyncJob) visitFile(path string, info os.FileInfo, err error) error {
	if err != nil {
		return fmt.Errorf("error walking the path %s, %w", path, err)
	}

	if info.IsDir() {
		if filesystem.IsIgnoredPath(path) {
			return filepath.SkipDir
		}
		tryIgnorePath := filepath.Join(path, ".gitignore")
		if _, err := os.Stat(tryIgnorePath); err == nil {
			err = j.ignoreHelper.AppendGitIgnoreFile(tryIgnorePath)
			if err != nil {
				return fmt.Errorf("AppendGitIgnoreFile: %w", err)
			}
		}
		matchResult, err := j.ignoreHelper.MatchsPath(path)
		if err != nil {
			return fmt.Errorf("MatchsPath: %w", err)
		}
		if matchResult {
			logrus.Println("ignore dir: ", path)
			return filepath.SkipDir
		}
	}
	logrus.Println("visitFile: ", path)

	matchResult, err := j.ignoreHelper.MatchsPath(path)
	if err != nil {
		return fmt.Errorf("MatchsPath: %w", err)
	}
	if matchResult {
		logrus.Println("ignore file: ", path)
		return nil
	}

	logrus.Println("matchResult: ", matchResult)

	var mimeType string
	var checksumValue string
	if !info.IsDir() {
		targetPath := strings.TrimPrefix(path, j.GitInfo.RootPath)
		targetPathWithRepo := fmt.Sprintf("%s%s%s%s", j.GitInfo.FirstCommitId, string(os.PathSeparator),
			j.GitInfo.Branch, targetPath)
		fullTargetPath, err := j.filePorter.CopyFile(path, targetPathWithRepo)
		if err != nil {
			logrus.Println("CopyFile: ", err)
			return fmt.Errorf("CopyFile: %w", err)
		}
		mimeType = helpers.GetMimeType(fullTargetPath)
		calcValue, err := checksum.CalcSha256(path)
		if err != nil {
			logrus.Println("CalcSha256: ", err)
			return fmt.Errorf("CalcSha256: %w", err)
		}
		checksumValue = calcValue
	}
	repoFile := &repo.MtRepoFileModel{
		Uid:        helpers.MustUuid(),
		Branch:     j.GitInfo.Branch,
		CommitId:   j.GitInfo.CommitId,
		SrcPath:    path,
		TargetPath: "fullTargetPath",
		Mime:       mimeType,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
		Checksum: sql.NullString{
			String: checksumValue,
			Valid:  true,
		},
		Syncno: sql.NullString{
			String: j.syncno,
			Valid:  true,
		},
		RepoId: sql.NullString{
			String: "", //j.GitInfo.RepoId,
			Valid:  false,
		},
		RepoFirstCommit: sql.NullString{
			String: j.GitInfo.FirstCommitId,
			Valid:  true,
		},
		RelativePath: sql.NullString{
			String: strings.ReplaceAll(strings.TrimPrefix(path, j.GitInfo.RootPath), string(os.PathSeparator), "/"),
			Valid:  true,
		},
		IsDir: sql.NullBool{
			Bool:  info.IsDir(),
			Valid: true,
		},
	}

	err = repo.PGInsertOrUpdateRepoFile(repoFile)
	if err != nil {
		logrus.Println("PGInsertOrUpdateRepoFile: ", err)
		return err
	}

	return nil
}

func (j *SyncJob) CopyFiles() (int, error) {

	err := filepath.Walk(j.RepoRootPath, j.visitFile)
	if err != nil {
		logrus.Printf("CopyFiles walking the path %v: %v\n", j.RepoRootPath, err)
		return 1, fmt.Errorf("CopyFiles walking the path %v: %v\n", j.RepoRootPath, err)
	}

	return 0, nil
}

type RepoWorker struct {
	repoChan   chan string
	repoJobMap map[string]*SyncJob
	filePorter *filesystem.FilePorter
	wg         *sync.WaitGroup
	syncno     string
}

func NewRepoWorker(wg *sync.WaitGroup, syncno string) (*RepoWorker, error) {
	storageUrl, ok := config.GetConfigurationString("STORAGE_URL")
	if !ok || storageUrl == "" {
		logrus.Fatalln("STORAGE_URL 未配置")
	}
	filePorter, err := filesystem.NewFilePorter(storageUrl)
	if err != nil {
		logrus.Errorln("初始化FilePorter失败", err)
		return nil, fmt.Errorf("初始化FilePorter失败")
	}

	return &RepoWorker{
		repoJobMap: make(map[string]*SyncJob),
		repoChan:   make(chan string, 2),
		filePorter: filePorter,
		wg:         wg,
		syncno:     syncno,
	}, nil
}

func (w *RepoWorker) AddJob(repoPath string) error {
	repoRoot, err := githelper.GitGetRepoRoot(repoPath)
	if err != nil {
		return fmt.Errorf("GitGetRepoRoot error: %s, %w", repoPath, err)
	}
	if repoRoot == "" {
		return fmt.Errorf("repoPath不能为空")
	}
	w.repoChan <- repoRoot
	return nil
}

func (w *RepoWorker) StartWork() {
	defer func() {
		logrus.Infoln("RepoWorker 退出")
		w.wg.Done()
	}()
	for {
		select {
		case repoPath := <-w.repoChan:
			if _, ok := w.repoJobMap[repoPath]; ok {
				continue
			}
			job, err := NewSyncJob(repoPath, w.filePorter, w.syncno)
			if err != nil {
				logrus.Fatalln("NewSyncJob: %w", err)
				return
			}
			w.repoJobMap[repoPath] = job
			w.wg.Add(1)
			go job.Sync(w.wg)
		}
	}
}
