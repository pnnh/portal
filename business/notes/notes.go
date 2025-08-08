package notes

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"neutron/services/strutil"

	"neutron/helpers"
	"neutron/services/datastore"

	"github.com/jmoiron/sqlx"
)

type MTNoteMatter struct {
	Cls         string `json:"cls"`
	Uid         string `json:"uid"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type MTNoteTable struct {
	Uid         string         `json:"uid"`
	Title       string         `json:"title"`
	Header      string         `json:"header"`
	Body        string         `json:"body"`
	Description string         `json:"description"`
	Keywords    string         `json:"keywords"`
	Status      int            `json:"status"`
	Cover       sql.NullString `json:"-"`
	Owner       string         `json:"-"`
	Channel     string         `json:"channel"`
	Discover    int            `json:"discover"`
	Partition   sql.NullString `json:"-"`
	CreateTime  time.Time      `json:"create_time" db:"create_time"`
	UpdateTime  time.Time      `json:"update_time" db:"update_time"`
	Version     sql.NullString `json:"-"`
	Build       sql.NullString `json:"-"`
	Url         sql.NullString `json:"-"`
	Branch      sql.NullString `json:"-"`
	Commit      sql.NullString `json:"-" db:"commit"`
	CommitTime  sql.NullTime   `json:"-" db:"commit_time"`
	RepoPath    sql.NullString `json:"-" db:"repo_path"`
	RepoId      sql.NullString `json:"-" db:"repo_id"`
	Cid         string         `json:"cid" db:"cid"`
	Lang        string         `json:"lang" db:"lang"`
	Nid         int64          `json:"nid" db:"nid" insert:"skip"`
	Dc          string         `json:"-" db:"dc"`
	Name        string         `json:"name" db:"name"`
}

func (t *MTNoteTable) ToModel() *MTNoteModel {
	return &MTNoteModel{
		MTNoteTable: *t,
		Partition:   t.Partition.String,
		Version:     t.Version.String,
		Build:       t.Build.String,
		Url:         t.Url.String,
		Branch:      t.Branch.String,
		Commit:      t.Commit.String,
		CommitTime:  t.CommitTime.Time,
		RepoPath:    t.RepoPath.String,
		RepoId:      t.RepoId.String,
	}
}

func (t *MTNoteTable) ToTableMap() (*datastore.TableMap, error) {
	tableMap := datastore.NewTableMap()
	tableMap.Set("uid", t.Uid)
	tableMap.Set("title", t.Title)
	tableMap.Set("header", t.Header)
	tableMap.Set("body", t.Body)
	tableMap.Set("description", t.Description)
	tableMap.Set("keywords", t.Keywords)
	tableMap.Set("status", t.Status)
	tableMap.Set("owner", t.Owner)
	tableMap.Set("channel", t.Channel)
	tableMap.Set("discover", t.Discover)
	tableMap.Set("create_time", t.CreateTime)
	tableMap.Set("update_time", t.UpdateTime)
	tableMap.Set("cid", t.Cid)
	tableMap.Set("lang", t.Lang)

	if t.Partition.Valid {
		tableMap.Set("partition", t.Partition)
	}
	if t.Version.Valid {
		tableMap.Set("version", t.Version)
	}
	if t.Build.Valid {
		tableMap.Set("build", t.Build)
	}
	if t.Url.Valid {
		tableMap.Set("url", t.Url)
	}
	if t.Branch.Valid {
		tableMap.Set("branch", t.Branch)
	}
	if t.Commit.Valid {
		tableMap.Set("commit", t.Commit)
	}
	if !t.CommitTime.Valid {
		tableMap.Set("commit_time", t.CommitTime)
	}
	if t.RepoPath.Valid {
		tableMap.Set("repo_path", t.RepoPath)
	}
	if t.RepoId.Valid {
		tableMap.Set("repo_id", t.RepoId)
	}

	return tableMap, nil

}

type MTNoteModel struct {
	MTNoteTable
	Partition  string    `json:"-"`
	Version    string    `json:"-"`
	Build      string    `json:"-"`
	Url        string    `json:"-"`
	Branch     string    `json:"-"`
	Commit     string    `json:"-" db:"commit"`
	CommitTime time.Time `json:"-" db:"commit_time"`
	RepoPath   string    `json:"-" db:"repo_path"`
	RepoId     string    `json:"-" db:"repo_id"`
}

func (m *MTNoteModel) ToViewModel() interface{} {
	view := &MTNoteView{
		MTNoteModel: *m,
	}
	//if m.Cid.Valid {
	//	view.Cid = m.Cid.String
	//}
	//if m.Lang.Valid {
	//	view.Lang = m.Lang.String
	//}
	view.Cid = m.Cid
	view.Lang = m.Lang
	return view

}

func (m *MTNoteModel) ToTableMap() (*datastore.TableMap, error) {
	tableMap := datastore.NewTableMap()
	tableMap.Set("uid", m.Uid)
	tableMap.Set("title", m.Title)
	tableMap.Set("header", m.Header)
	tableMap.Set("body", m.Body)
	tableMap.Set("description", m.Description)
	tableMap.Set("keywords", m.Keywords)
	tableMap.Set("status", m.Status)
	tableMap.Set("owner", m.Owner)
	tableMap.Set("channel", m.Channel)
	tableMap.Set("discover", m.Discover)
	tableMap.Set("create_time", m.CreateTime)
	tableMap.Set("update_time", m.UpdateTime)
	tableMap.Set("cid", m.Cid)
	tableMap.Set("lang", m.Lang)

	if m.Partition != "" {
		tableMap.Set("partition", m.Partition)
	}
	if m.Version != "" {
		tableMap.Set("version", m.Version)
	}
	if m.Build != "" {
		tableMap.Set("build", m.Build)
	}
	if m.Url != "" {
		tableMap.Set("url", m.Url)
	}
	if m.Branch != "" {
		tableMap.Set("branch", m.Branch)
	}
	if m.Commit != "" {
		tableMap.Set("commit", m.Commit)
	}
	if !m.CommitTime.IsZero() {
		tableMap.Set("commit_time", m.CommitTime)
	}
	if m.RepoPath != "" {
		tableMap.Set("repo_path", m.RepoPath)
	}
	if m.RepoId != "" {
		tableMap.Set("repo_id", m.RepoId)
	}

	return tableMap, nil

}

type MTNoteView struct {
	MTNoteModel
	Cid  string `json:"cid" db:"cid"`
	Lang string `json:"lang" db:"lang"`
}

func (v *MTNoteView) ToModel() *MTNoteModel {
	return &MTNoteModel{
		MTNoteTable: v.MTNoteModel.MTNoteTable,
		Partition:   v.Partition,
		Version:     v.Version,
		Build:       v.Build,
		Url:         v.Url,
		Branch:      v.Branch,
		Commit:      v.Commit,
		CommitTime:  v.CommitTime,
		RepoPath:    v.RepoPath,
		RepoId:      v.RepoId,
	}
}

func PGConsoleInsertNote(tableMapConverter datastore.IConvertTableMap) error {
	tableMap, err := tableMapConverter.ToTableMap()
	if err != nil {
		return fmt.Errorf("PGConsoleInsertNote ToTableMap: %w", err)
	}

	//columnMap, err := datastore.ReflectColumns(&model.MTNoteTable)
	//if err != nil {
	//	return fmt.Errorf("PGConsoleInsertNote ReflectColumns: %w", err)
	//}

	colNames := tableMap.Keys()
	colText := strings.Join(colNames, ", ")
	colPlaceholders := strutil.JoinStringsFunc(colNames, func(s string) string {
		return fmt.Sprintf(":%s, ", s)
	}, func(s string) string {
		return strings.TrimRight(s, ", ")
	})

	sqlText := fmt.Sprintf(`insert into articles(%s)
values(%s)
on conflict (uid)
do nothing;`, colText, colPlaceholders)

	_, err = datastore.NamedExec(sqlText, tableMap)
	if err != nil {
		return fmt.Errorf("PGConsoleInsertNote: %w", err)
	}
	return nil
}

func SelectNotes(channel, keyword string, page int, size int, lang string) (*helpers.Pagination, []*MTNoteTable, error) {
	pagination := helpers.CalcPaginationByPage(page, size)
	baseSqlText := ` select * from articles `
	baseSqlParams := map[string]interface{}{}

	whereText := ` where status = 1 `
	if keyword != "" {
		whereText += ` and (title like :keyword or description like :keyword) `
		baseSqlParams["keyword"] = "%" + keyword + "%"
	}
	if channel != "" {
		whereText += ` and channel = :channel `
		baseSqlParams["channel"] = channel
	}
	if lang != "" {
		whereText += ` and lang = :lang `
		baseSqlParams["lang"] = lang
	}
	orderText := ` order by create_time desc `

	pageSqlText := fmt.Sprintf("%s %s %s %s", baseSqlText, whereText, orderText, ` offset :offset limit :limit; `)
	pageSqlParams := map[string]interface{}{
		"offset": pagination.Offset, "limit": pagination.Limit,
	}
	for k, v := range baseSqlParams {
		pageSqlParams[k] = v
	}
	var sqlResults []*MTNoteTable

	rows, err := datastore.NamedQuery(pageSqlText, pageSqlParams)
	if err != nil {
		return nil, nil, fmt.Errorf("NamedQuery: %w", err)
	}
	if err = sqlx.StructScan(rows, &sqlResults); err != nil {
		return nil, nil, fmt.Errorf("StructScan: %w", err)
	}

	countSqlText := `select count(1) as count from (` +
		fmt.Sprintf("%s %s", baseSqlText, whereText) + `) as temp;`

	countSqlParams := map[string]interface{}{}
	for k, v := range baseSqlParams {
		countSqlParams[k] = v
	}
	var countSqlResults []struct {
		Count int `db:"count"`
	}

	rows, err = datastore.NamedQuery(countSqlText, countSqlParams)
	if err != nil {
		return nil, nil, fmt.Errorf("NamedQuery: %w", err)
	}
	if err = sqlx.StructScan(rows, &countSqlResults); err != nil {
		return nil, nil, fmt.Errorf("StructScan: %w", err)
	}
	if len(countSqlResults) == 0 {
		return nil, nil, fmt.Errorf("查询笔记总数有误，数据为空")
	}

	//selectData := &nemodels.NESelectResult[*MTNoteModel]{
	//	Page:  pagination.Page,
	//	Size:  pagination.Size,
	//	Count: countSqlResults[0].Count,
	//	Range: resultRange,
	//}

	return pagination, sqlResults, nil
}

// PGGetNote 获取单个笔记信息
// obsolete 处理向后兼容逻辑。早期是通过单个uid查询，后期可以通过cid和lang查询。
func PGGetNote(uid string, lang string) (*MTNoteTable, error) {
	if uid == "" {
		return nil, fmt.Errorf("PGGetNote uid is empty")
	}
	pageSqlText := ` select * from articles where status = 1 and (uid = :uid or (cid = :uid and lang = :lang)); `

	pageSqlParams := map[string]interface{}{
		"uid":  uid,
		"lang": lang,
	}
	var sqlResults []*MTNoteTable

	rows, err := datastore.NamedQuery(pageSqlText, pageSqlParams)
	if err != nil {
		return nil, fmt.Errorf("NamedQuery: %w", err)
	}
	if err = sqlx.StructScan(rows, &sqlResults); err != nil {
		return nil, fmt.Errorf("StructScan: %w", err)
	}

	for _, item := range sqlResults {
		return item, nil
	}
	return nil, nil
}

type MTNoteFileModel struct {
	Title       string `json:"title"`
	Path        string `json:"path"`
	IsDir       bool   `json:"is_dir"`
	IsText      bool   `json:"is_text"`
	IsImage     bool   `json:"is_image"`
	StoragePath string `json:"storage_path"`
}
