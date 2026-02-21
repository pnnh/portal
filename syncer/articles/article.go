package articles

import (
	"fmt"
	"os"
	"path/filepath"
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

type ArticleWorker struct {
	repoWorker *RepoWorker
	rootPath   string
	syncno     string
	dirUidMap  map[string]string
}

func NewArticleWorker(repoWorker *RepoWorker, rootPath string, syncno string) (*ArticleWorker, error) {
	return &ArticleWorker{
		repoWorker: repoWorker,
		rootPath:   rootPath,
		syncno:     syncno,
		dirUidMap:  make(map[string]string),
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

	fileName := strings.ToLower(filepath.Base(path))
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
		w.dirUidMap[path] = newUid
	} else {
		sum, err := checksum.CalcSha256(path)
		if err != nil {
			return fmt.Errorf("计算文件校验和失败: %w", err)
		}
		sumValue = sum
		mimeType = helpers.GetMimeType(path)
		parentDir := filepath.Dir(path)
		parentUid = w.dirUidMap[parentDir]
	}

	logrus.Infoln("开始同步文章: ", path)

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
	dataRow.SetString("url", "")
	dataRow.SetNullString("parent", parentUid)
	dataRow.SetNullString("name", fileName)
	dataRow.SetNullString("checksum", sumValue)
	dataRow.SetNullString("syncno", w.syncno)
	dataRow.SetNullString("mimetype", mimeType)

	err = PGInsertFile(dataRow)
	if err != nil {
		logrus.Errorf("插入文件数据失败: %v", err)
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
