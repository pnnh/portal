package notes

import (
	"database/sql"
	"fmt"
	"time"

	"neutron/helpers"
	"neutron/services/datastore"

	"github.com/jmoiron/sqlx"
)

type MTNotetransTable struct {
	Uid         string         `json:"uid"`
	Title       string         `json:"title"`
	Header      string         `json:"header"`
	Body        string         `json:"body"`
	Description string         `json:"description"`
	Keywords    string         `json:"keywords"`
	Status      int            `json:"status"`
	Cover       sql.NullString `json:"-"`
	Owner       string         `json:"-"`
	Channel     sql.NullString `json:"channel"`
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
	Lang        string         `json:"lang" db:"lang"`
	Name        string         `json:"name" db:"name"`
}

func (t *MTNotetransTable) ToModel() *MTNotetransModel {
	return &MTNotetransModel{
		MTNotetransTable: *t,
		Partition:        t.Partition.String,
		Version:          t.Version.String,
		Build:            t.Build.String,
		Url:              t.Url.String,
		Branch:           t.Branch.String,
		Commit:           t.Commit.String,
		CommitTime:       t.CommitTime.Time,
		RepoPath:         t.RepoPath.String,
		RepoId:           t.RepoId.String,
		Channel:          t.Channel.String,
	}
}

//func (t *MTNotetransTable) ToTableMap() (*datastore.DataRow, error) {
//	tableMap := datastore.NewDataRow()
//	tableMap.Set("uid", t.Uid)
//	tableMap.Set("title", t.Title)
//	tableMap.Set("header", t.Header)
//	tableMap.Set("body", t.Body)
//	tableMap.Set("description", t.Description)
//	tableMap.Set("keywords", t.Keywords)
//	tableMap.Set("status", t.Status)
//	tableMap.Set("owner", t.Owner)
//	tableMap.Set("channel", t.Channel)
//	tableMap.Set("discover", t.Discover)
//	tableMap.Set("create_time", t.CreateTime)
//	tableMap.Set("update_time", t.UpdateTime)
//	tableMap.Set("lang", t.Lang)
//
//	if t.Partition.Valid {
//		tableMap.Set("partition", t.Partition.String)
//	}
//	if t.Version.Valid {
//		tableMap.Set("version", t.Version.String)
//	}
//	if t.Build.Valid {
//		tableMap.Set("build", t.Build.String)
//	}
//	if t.Url.Valid {
//		tableMap.Set("url", t.Url.String)
//	}
//	if t.Branch.Valid {
//		tableMap.Set("branch", t.Branch.String)
//	}
//	if t.Commit.Valid {
//		tableMap.Set("commit", t.Commit.String)
//	}
//	if !t.CommitTime.Valid {
//		tableMap.Set("commit_time", t.CommitTime.Time)
//	}
//	if t.RepoPath.Valid {
//		tableMap.Set("repo_path", t.RepoPath.String)
//	}
//	if t.RepoId.Valid {
//		tableMap.Set("repo_id", t.RepoId.String)
//	}
//
//	return tableMap, nil
//
//}

type MTNotetransModel struct {
	MTNotetransTable
	Partition  string    `json:"-"`
	Version    string    `json:"-"`
	Build      string    `json:"-"`
	Url        string    `json:"-"`
	Branch     string    `json:"-"`
	Commit     string    `json:"-" db:"commit"`
	CommitTime time.Time `json:"-" db:"commit_time"`
	RepoPath   string    `json:"-" db:"repo_path"`
	RepoId     string    `json:"-" db:"repo_id"`
	Channel    string    `json:"-" db:"channel"`
}

func (m *MTNotetransModel) ToViewModel() interface{} {
	view := &MTNotetransView{
		MTNotetransModel: *m,
	}
	//if m.Lang.Valid {
	//	view.Lang = m.Lang.String
	//}
	view.Lang = m.Lang
	return view

}

//func (m *MTNotetransModel) ToTableMap() (*datastore.DataRow, error) {
//	tableMap := datastore.NewDataRow()
//	tableMap.Set("uid", m.Uid)
//	tableMap.Set("title", m.Title)
//	tableMap.Set("header", m.Header)
//	tableMap.Set("body", m.Body)
//	tableMap.Set("description", m.Description)
//	tableMap.Set("keywords", m.Keywords)
//	tableMap.Set("status", m.Status)
//	tableMap.Set("owner", m.Owner)
//	tableMap.Set("discover", m.Discover)
//	tableMap.Set("create_time", m.CreateTime)
//	tableMap.Set("update_time", m.UpdateTime)
//	tableMap.Set("lang", m.Lang)
//
//	if m.Partition != "" {
//		tableMap.Set("partition", m.Partition)
//	}
//	if m.Version != "" {
//		tableMap.Set("version", m.Version)
//	}
//	if m.Build != "" {
//		tableMap.Set("build", m.Build)
//	}
//	if m.Url != "" {
//		tableMap.Set("url", m.Url)
//	}
//	if m.Branch != "" {
//		tableMap.Set("branch", m.Branch)
//	}
//	if m.Commit != "" {
//		tableMap.Set("commit", m.Commit)
//	}
//	if !m.CommitTime.IsZero() {
//		tableMap.Set("commit_time", m.CommitTime)
//	}
//	if m.RepoPath != "" {
//		tableMap.Set("repo_path", m.RepoPath)
//	}
//	if m.RepoId != "" {
//		tableMap.Set("repo_id", m.RepoId)
//	}
//	if m.Channel != "" {
//		tableMap.Set("channel", m.Channel)
//	}
//
//	return tableMap, nil
//
//}

type MTNotetransView struct {
	MTNotetransModel
	Lang string `json:"lang" db:"lang"`
}

func (v *MTNotetransView) ToModel() *MTNotetransModel {
	return &MTNotetransModel{
		MTNotetransTable: v.MTNotetransModel.MTNotetransTable,
		Partition:        v.Partition,
		Version:          v.Version,
		Build:            v.Build,
		Url:              v.Url,
		Branch:           v.Branch,
		Commit:           v.Commit,
		CommitTime:       v.CommitTime,
		RepoPath:         v.RepoPath,
		RepoId:           v.RepoId,
	}
}

//func PGConsoleInsertNotetrans(tableMapConverter datastore.IConvertTableMap) error {
//tableMap, err := tableMapConverter.ToTableMap()
//if err != nil {
//	return fmt.Errorf("PGConsoleInsertNotetrans ToTableMap: %w", err)
//}

//columnMap, err := datastore.ReflectColumns(&model.MTNotetransTable)
//if err != nil {
//	return fmt.Errorf("PGConsoleInsertNotetrans ReflectColumns: %w", err)
//}

//	colNames := tableMap.Keys()
//	if len(colNames) == 0 {
//		return fmt.Errorf("PGConsoleInsertNotetrans: no columns to insert")
//	}
//	colText := strings.Join(colNames, ", ")
//	colPlaceholders := strutil.JoinStringsFunc(colNames, func(s string) string {
//		return fmt.Sprintf(":%s, ", s)
//	}, func(s string) string {
//		return strings.TrimRight(s, ", ")
//	})
//
//	sqlText := fmt.Sprintf(`insert into articles(%s)
//values(%s)
//on conflict (uid)
//do nothing;`, colText, colPlaceholders)
//
//	paramsMap := tableMap.MapData()
//
//	_, err = datastore.NamedExec(sqlText, paramsMap)
//	if err != nil {
//		return fmt.Errorf("PGConsoleInsertNotetrans: %w", err)
//	}
//	return nil
//}

func SelectNotetranss(channel, keyword string, page int, size int, lang string) (*helpers.Pagination, []*MTNotetransTable, error) {
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
	//if lang != "" {
	//	whereText += ` and lang = :lang `
	//	baseSqlParams["lang"] = lang
	//}
	orderText := ` order by create_time desc `

	pageSqlText := fmt.Sprintf("%s %s %s %s", baseSqlText, whereText, orderText, ` offset :offset limit :limit; `)
	pageSqlParams := map[string]interface{}{
		"offset": pagination.Offset, "limit": pagination.Limit,
	}
	for k, v := range baseSqlParams {
		pageSqlParams[k] = v
	}
	var sqlResults []*MTNotetransTable

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

	//selectData := &nemodels.NESelectResult[*MTNotetransModel]{
	//	Page:  pagination.Page,
	//	Size:  pagination.Size,
	//	Count: countSqlResults[0].Count,
	//	Range: resultRange,
	//}

	return pagination, sqlResults, nil
}

// PGGetNotetrans 获取单个笔记信息
//func PGGetNotetrans(uid string, lang string) (*MTNotetransTable, error) {
//	if uid == "" {
//		return nil, fmt.Errorf("PGGetNotetrans uid is empty")
//	}
//	pageSqlText := ` select * from articles where status = 1 and (uid = :uid or (cid = :uid and lang = :lang)); `
//
//	pageSqlParams := map[string]interface{}{
//		"uid":  uid,
//		"lang": lang,
//	}
//	var sqlResults []*MTNotetransTable
//
//	rows, err := datastore.NamedQuery(pageSqlText, pageSqlParams)
//	if err != nil {
//		return nil, fmt.Errorf("NamedQuery: %w", err)
//	}
//	if err = sqlx.StructScan(rows, &sqlResults); err != nil {
//		return nil, fmt.Errorf("StructScan: %w", err)
//	}
//
//	for _, item := range sqlResults {
//		return item, nil
//	}
//	return nil, nil
//}
