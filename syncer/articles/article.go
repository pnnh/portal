package articles

import (
	"fmt"
	"os"
	"path/filepath"
	"portal/services/base58"
	"strings"
	"time"

	"github.com/pnnh/neutron/services/filesystem"

	"github.com/pnnh/neutron/services/checksum"
	"github.com/pnnh/neutron/services/datastore"

	"github.com/pnnh/neutron/helpers"
	"github.com/sirupsen/logrus"
)

// 一个固定的用户ID，通过Syncer服务同步的文章属于这个用户
const SyncerArticleOwner = "01990e6a-2689-731b-a5a2-b46117e22040"
const SyncParentUid = "76de121c-0fab-11f1-a643-6c02e0549f86"

type dirStat struct {
	uid    string
	synced bool
}

type ArticleWorker struct {
	repoWorker *RepoWorker
	rootPath   string
	syncno     string
	dirStatMap map[string]*dirStat
}

func NewArticleWorker(repoWorker *RepoWorker, rootPath string, syncno string) (*ArticleWorker, error) {
	return &ArticleWorker{
		repoWorker: repoWorker,
		rootPath:   rootPath,
		syncno:     syncno,
		dirStatMap: make(map[string]*dirStat),
	}, nil
}

func (w *ArticleWorker) StartWork() {
	err := filepath.Walk(w.rootPath, w.visitFile)
	if err != nil {
		logrus.Fatalln("error walking the path", w.rootPath, err)
	}
}

func (w *ArticleWorker) visitFile(path string, info os.FileInfo, visitErr error) error {
	if visitErr != nil {
		return fmt.Errorf("error walking the path %s, %w", path, visitErr)
	}

	fileName := filepath.Base(path)
	if info.IsDir() && strings.HasPrefix(fileName, ".") {
		return filepath.SkipDir
	}
	if filesystem.IsIgnoredPath(path) {
		if info.IsDir() {
			return filepath.SkipDir
		}
		// 匹配到忽略的文件时继续遍历当前目录下的其它文件或目录
		return nil
	}
	sumValue := ""
	mimeType := ""
	newUid := helpers.MustUuid()
	parentUid := ""
	if info.IsDir() {
		mimeType = "directory"
		if path != w.rootPath {
			w.dirStatMap[path] = &dirStat{uid: newUid, synced: false}
		}
	} else {
		sum, err := checksum.CalcSha256(path)
		if err != nil {
			return fmt.Errorf("计算文件校验和失败: %w", err)
		}
		sumValue = sum
		mimeType = helpers.GetMimeType(path)
	}
	parentDir := filepath.Dir(path)
	parentDirStat := w.dirStatMap[parentDir]
	if parentDirStat != nil {
		if !parentDirStat.synced {
			return filepath.SkipDir
		}
		parentUid = parentDirStat.uid
	}

	// 执行到此处parentUid为空说明处于根目录下，根目录的父目录UID设置为一个固定值，表示没有父目录
	if parentUid == "" {
		if parentDir == w.rootPath {
			parentUid = SyncParentUid
		} else if path == w.rootPath {
			// 根目录不处理，直接返回
			return nil
		} else {
			panic(fmt.Sprintf("未找到父目录UID: %s, parentDir: %s", parentUid, parentDir))
		}
	}

	noteTitle := strings.Trim(fileName, " \n\r\t ")
	if noteTitle == "" {
		noteTitle = strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))
	}

	nowTime := time.Now()
	dataRow := datastore.NewDataRow()
	dataRow.SetString("uid", newUid)
	dataRow.SetString("title", noteTitle)
	dataRow.SetString("header", "{}")
	dataRow.SetString("body", "{}")
	dataRow.SetString("description", "")
	dataRow.SetString("keywords", "")
	dataRow.SetInt("status", 1)
	dataRow.SetNullString("cover", "")
	dataRow.SetString("owner", SyncerArticleOwner)
	dataRow.SetNullString("channel", "")
	dataRow.SetInt("discover", 0)
	dataRow.SetNullString("partition", "")
	dataRow.SetTime("create_time", nowTime)
	dataRow.SetTime("update_time", nowTime)
	dataRow.SetNullString("version", "0")
	dataRow.SetString("lang", "")
	dataRow.SetNullString("parent", parentUid)
	dataRow.SetNullString("name", fileName)
	dataRow.SetNullString("checksum", sumValue)
	dataRow.SetNullString("syncno", w.syncno)
	dataRow.SetNullString("mimetype", mimeType)
	dataRow.SetString("url", "")

	targetPath := ""
	if !info.IsDir() {
		logrus.Infoln("同步文件: ", path, "，checksum: ", sumValue)

		targetParentDir, err := base58.UuidToBase58(parentUid)
		if err != nil {
			logrus.Errorf("转换父目录UID失败: %v", err)
			return nil
		}
		targetSelfName, err := base58.UuidToBase58(newUid)
		if err != nil {
			logrus.Errorf("转换文件UID失败: %v", err)
			return nil
		}
		if !strings.HasPrefix(fileName, ".") {
			extName := filepath.Ext(fileName)
			if extName != "" {
				targetSelfName += extName
			}
		}
		targetPath = fmt.Sprintf("%s/%s", targetParentDir, targetSelfName)
		targetUrl := fmt.Sprintf("storage://%s", targetPath)
		dataRow.SetString("url", targetUrl)
	}

	// 根目录本身不插入数据库记录，直接遍历其下的文件或目录
	if path == w.rootPath {
		return nil
	}
	err := PGInsertFile(dataRow)
	if err != nil {
		logrus.Errorf("插入文件数据失败: %s %v", fileName, err)
		return nil
	}

	if !info.IsDir() {
		w.repoWorker.AddJob(path, targetPath)
	} else {
		currentDirStat := w.dirStatMap[path]
		if currentDirStat != nil {
			currentDirStat.synced = true
		}
	}

	return nil
}

func PGInsertFile(dataRow *datastore.DataRow) error {
	sqlText := `insert into files(uid, title, header, body, create_time, update_time, keywords, description, status, 
	cover, owner, discover, version, url, 
	lang, name, checksum, syncno, mimetype, parent)
values(:uid, :title, :header, :body, :create_time, :update_time, :keywords, :description, :status, :cover, :owner, 
	:discover, :version, :url, 
	:lang, :name, :checksum, :syncno, :mimetype, :parent);`

	paramsMap := dataRow.InnerMap()

	_, err := datastore.NamedExec(sqlText, paramsMap)
	if err != nil {
		return fmt.Errorf("PGConsoleInsertNote: %w", err)
	}
	return nil
}
