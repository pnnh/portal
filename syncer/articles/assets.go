package articles

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"neutron/config"
	"neutron/helpers"
	"neutron/services/filesystem"
	"portal/models/repo"
	"portal/services/githelper"
)

type SyncJob struct {
	RepoRootPath string
	GitInfo      *githelper.GitInfo
	filePorter   *filesystem.FilePorter
	ignoreHelper *githelper.GitIgnoreHelper
}

func NewSyncJob(repoPath string, filePorter *filesystem.FilePorter) (*SyncJob, error) {
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
	}, nil
}

func (j *SyncJob) Sync() {
	// 如果当前commit不是clean状态，则不需要同步
	if !j.GitInfo.IsClean {
		logrus.Println("工作区不干净跳过同步: ", j.RepoRootPath)
		return
	}
	repoSyncInfo, err := repo.PGGetRepoSyncInfo(j.GitInfo.RepoId, j.GitInfo.Branch)
	if err != nil {
		logrus.Fatalln("获取repo sync info失败: ", j.RepoRootPath, err)
		return
	}
	if repoSyncInfo != nil && repoSyncInfo.LastCommitId != "" {
		isAncestor, err := githelper.GitCommitIsAncestor(j.RepoRootPath, repoSyncInfo.LastCommitId, j.GitInfo.CommitId)
		if err != nil {
			logrus.Println("比较commit失败: ", j.RepoRootPath, err)
			return
		}
		// 如果当前commit在上次commit之前，则不需要同步
		if !isAncestor {
			logrus.Println("状态已最新无需同步: ", j.RepoRootPath)
			return
		}
	}
	_, err = j.CopyFiles()
	if err != nil {
		logrus.Fatalln("CopyFiles: ", err)
		return
	}
	repoSyncInfo = &repo.MTRepoSyncModel{
		Uid:          helpers.MustUuid(),
		LastCommitId: j.GitInfo.CommitId,
		Branch:       j.GitInfo.Branch,
		RepoId:       j.GitInfo.RepoId,
	}
	err = repo.PGInsertOrUpdateRepoSyncInfo(repoSyncInfo)
	if err != nil {
		logrus.Fatalln("PGInsertOrUpdateRepoSyncInfo: ", err)
		return
	}
}

func (j *SyncJob) visitFile(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if info.IsDir() {

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

		return nil
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
	targetPath := strings.TrimPrefix(path, j.GitInfo.RootPath)
	targetPath = fmt.Sprintf("%s/%s%s", j.GitInfo.RepoId, j.GitInfo.Branch, targetPath)
	fullTargetPath, err := j.filePorter.CopyFile(path, targetPath)
	if err != nil {
		logrus.Println("CopyFile: ", err)
		return fmt.Errorf("CopyFile: %w", err)
	}

	repoFile := &repo.MtRepoFileModel{
		Uid:        helpers.MustUuid(),
		Branch:     j.GitInfo.Branch,
		CommitId:   j.GitInfo.CommitId,
		SrcPath:    path,
		TargetPath: fullTargetPath,
		Mime:       "",
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
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
		fmt.Printf("CopyFiles walking the path %v: %v\n", j.RepoRootPath, err)
		return 1, fmt.Errorf("CopyFiles walking the path %v: %v\n", j.RepoRootPath, err)
	}

	return 0, nil
}

type RepoWorker struct {
	repoChan   chan string
	repoJobMap map[string]*SyncJob
	filePorter *filesystem.FilePorter
}

func NewRepoWorker() (*RepoWorker, error) {
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
	}, nil
}

func (w *RepoWorker) AddJob(repoPath string) error {
	repoRoot, err := githelper.GitGetRepoRoot(repoPath)
	if err != nil {
		return fmt.Errorf("GitGetRepoRoot: %w", err)
	}
	if repoRoot == "" {
		return fmt.Errorf("repoPath不能为空")
	}
	w.repoChan <- repoRoot
	return nil
}

func (w *RepoWorker) StartWork() {
	for {
		select {
		case repoPath := <-w.repoChan:
			if _, ok := w.repoJobMap[repoPath]; ok {
				continue
			}
			job, err := NewSyncJob(repoPath, w.filePorter)
			if err != nil {
				logrus.Fatalln("NewSyncJob: %w", err)
				return
			}
			w.repoJobMap[repoPath] = job
			go job.Sync()
		}
	}
}
