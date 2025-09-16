package notes

import (
	"database/sql"
	"fmt"
	"time"

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
	Uid             string         `json:"uid"`
	Title           string         `json:"title"`
	Header          string         `json:"header"`
	Body            string         `json:"body"`
	Description     string         `json:"description"`
	Keywords        string         `json:"keywords"`
	Status          int            `json:"status"`
	Cover           sql.NullString `json:"-"`
	Owner           string         `json:"-"`
	Channel         sql.NullString `json:"channel"`
	Discover        int            `json:"discover"`
	Partition       sql.NullString `json:"-"`
	CreateTime      time.Time      `json:"create_time" db:"create_time"`
	UpdateTime      time.Time      `json:"update_time" db:"update_time"`
	Version         sql.NullString `json:"-"`
	Build           sql.NullString `json:"-"`
	Url             sql.NullString `json:"-"`
	Branch          sql.NullString `json:"-"`
	Commit          sql.NullString `json:"-" db:"commit"`
	CommitTime      sql.NullTime   `json:"-" db:"commit_time"`
	RelativePath    sql.NullString `json:"-" db:"relative_path"`
	RepoId          sql.NullString `json:"-" db:"repo_id"`
	Lang            string         `json:"lang" db:"lang"`
	Name            string         `json:"name" db:"name"`
	Checksum        sql.NullString `json:"checksum" db:"checksum"` // 用于标识内容是否变更
	Syncno          sql.NullString `json:"-" db:"syncno"`          // 用于标识同步批次
	RepoFirstCommit sql.NullString `json:"-" db:"repo_first_commit"`
}

func (t *MTNoteTable) ToModel() *MTNoteModel {
	return &MTNoteModel{
		MTNoteTable:     *t,
		Partition:       t.Partition.String,
		Version:         t.Version.String,
		Build:           t.Build.String,
		Url:             t.Url.String,
		Branch:          t.Branch.String,
		Commit:          t.Commit.String,
		CommitTime:      t.CommitTime.Time,
		RelativePath:    t.RelativePath.String,
		RepoId:          t.RepoId.String,
		Channel:         t.Channel.String,
		CheckSum:        t.Checksum.String,
		Syncno:          t.Syncno.String,
		RepoFirstCommit: t.RepoFirstCommit.String,
	}
}

//
//func (t *MTNoteTable) ToTableMap() (*datastore.TableMap, error) {
//	tableMap := datastore.NewTableMap()
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
//	if t.RelativePath.Valid {
//		tableMap.Set("repo_path", t.RelativePath.String)
//	}
//	if t.RepoId.Valid {
//		tableMap.Set("repo_id", t.RepoId.String)
//	}
//	if t.Checksum.Valid {
//		tableMap.Set("checksum", t.Checksum.String)
//	}
//	if t.Syncno.Valid {
//		tableMap.Set("syncno", t.Syncno.String)
//	}
//
//	return tableMap, nil
//
//}

type MTNoteModel struct {
	MTNoteTable
	Partition       string    `json:"partition"`
	Version         string    `json:"version"`
	Build           string    `json:"build"`
	Url             string    `json:"url"`
	Branch          string    `json:"branch"`
	Commit          string    `json:"commit" db:"commit"`
	CommitTime      time.Time `json:"commit_time" db:"commit_time"`
	RelativePath    string    `json:"relative_path" db:"relative_path"`
	RepoId          string    `json:"repo_id" db:"repo_id"`
	Channel         string    `json:"channel" db:"channel"`
	Cover           string    `json:"cover" db:"cover"`
	CheckSum        string    `json:"checksum" db:"checksum"`
	Syncno          string    `json:"syncno" db:"syncno"`
	RepoFirstCommit string    `json:"-" db:"repo_first_commit"`
}

func (m *MTNoteModel) ToViewModel() interface{} {
	view := &MTNoteView{
		MTNoteModel: *m,
	}
	//if m.Lang.Valid {
	//	view.Lang = m.Lang.String
	//}
	view.Lang = m.Lang
	return view

}

func (m *MTNoteModel) ToTableMap() (*MTNoteTable, error) {
	//tableMap := datastore.NewTableMap()
	//tableMap.Set("uid", m.Uid)
	//tableMap.Set("title", m.Title)
	//tableMap.Set("header", m.Header)
	//tableMap.Set("body", m.Body)
	//tableMap.Set("description", m.Description)
	//tableMap.Set("keywords", m.Keywords)
	//tableMap.Set("status", m.Status)
	//tableMap.Set("owner", m.Owner)
	//tableMap.Set("discover", m.Discover)
	//tableMap.Set("create_time", m.CreateTime)
	//tableMap.Set("update_time", m.UpdateTime)
	//tableMap.Set("lang", m.Lang)
	//
	//if m.Partition != "" {
	//	tableMap.Set("partition", m.Partition)
	//}
	//if m.Version != "" {
	//	tableMap.Set("version", m.Version)
	//}
	//if m.Build != "" {
	//	tableMap.Set("build", m.Build)
	//}
	//if m.Url != "" {
	//	tableMap.Set("url", m.Url)
	//}
	//if m.Branch != "" {
	//	tableMap.Set("branch", m.Branch)
	//}
	//if m.Commit != "" {
	//	tableMap.Set("commit", m.Commit)
	//}
	//if !m.CommitTime.IsZero() {
	//	tableMap.Set("commit_time", m.CommitTime)
	//}
	//if m.RelativePath != "" {
	//	tableMap.Set("repo_path", m.RelativePath)
	//}
	//if m.RepoId != "" {
	//	tableMap.Set("repo_id", m.RepoId)
	//}
	//if m.Channel != "" {
	//	tableMap.Set("channel", m.Channel)
	//}
	//
	//return tableMap, nil
	table := &MTNoteTable{
		Uid:          m.Uid,
		Title:        m.Title,
		Header:       m.Header,
		Body:         m.Body,
		Description:  m.Description,
		Keywords:     m.Keywords,
		Status:       m.Status,
		Cover:        sql.NullString{String: m.Cover, Valid: true},
		Owner:        m.Owner,
		Channel:      sql.NullString{String: m.Channel, Valid: true},
		Discover:     m.Discover,
		Partition:    sql.NullString{String: m.Partition, Valid: true},
		CreateTime:   m.CreateTime,
		UpdateTime:   m.UpdateTime,
		Version:      sql.NullString{String: m.Version, Valid: true},
		Build:        sql.NullString{String: m.Build, Valid: true},
		Url:          sql.NullString{String: m.Url, Valid: true},
		Branch:       sql.NullString{String: m.Branch, Valid: true},
		Commit:       sql.NullString{String: m.Commit, Valid: true},
		CommitTime:   sql.NullTime{Time: m.CommitTime, Valid: true},
		RelativePath: sql.NullString{String: m.RelativePath, Valid: true},
		RepoId:       sql.NullString{String: m.RepoId, Valid: true},
		Lang:         m.Lang,
		Name:         m.Name,
		Checksum:     sql.NullString{String: m.CheckSum, Valid: true},
		Syncno:       sql.NullString{String: m.Syncno, Valid: true},
	}
	return table, nil
}

type MTNoteView struct {
	MTNoteModel
	Lang string `json:"lang" db:"lang"`
}

func (v *MTNoteView) ToModel() *MTNoteModel {
	return &MTNoteModel{
		MTNoteTable:  v.MTNoteModel.MTNoteTable,
		Partition:    v.Partition,
		Version:      v.Version,
		Build:        v.Build,
		Url:          v.Url,
		Branch:       v.Branch,
		Commit:       v.Commit,
		CommitTime:   v.CommitTime,
		RelativePath: v.RelativePath,
		RepoId:       v.RepoId,
	}
}

func PGConsoleInsertNote(noteTable *MTNoteTable) error {
	//tableMap, err := tableMapConverter.ToTableMap()
	//if err != nil {
	//	return fmt.Errorf("PGConsoleInsertNote ToTableMap: %w", err)
	//}

	//columnMap, err := datastore.ReflectColumns(&model.MTNoteTable)
	//if err != nil {
	//	return fmt.Errorf("PGConsoleInsertNote ReflectColumns: %w", err)
	//}

	//colNames := tableMap.Keys()
	//if len(colNames) == 0 {
	//	return fmt.Errorf("PGConsoleInsertNote: no columns to insert")
	//}
	//colText := strings.Join(colNames, ", ")
	//colPlaceholders := strutil.JoinStringsFunc(colNames, func(s string) string {
	//	return fmt.Sprintf(":%s, ", s)
	//}, func(s string) string {
	//	return strings.TrimRight(s, ", ")
	//})

	sqlText := `insert into articles(uid, title, header, body, create_time, update_time, keywords, description, status, 
	cover, owner, channel, discover, partition, version, build, url, branch, commit, commit_time, relative_path, repo_id, 
	lang, name, checksum, syncno, repo_first_commit)
values(:uid, :title, :header, :body, :create_time, :update_time, :keywords, :description, :status, :cover, :owner, 
	:channel, :discover, :partition, :version, :build, :url, :branch, :commit, :commit_time, :relative_path, :repo_id, 
	:lang, :name, :checksum, :syncno, :repo_first_commit)
on conflict (uid)
do update set title = excluded.title,
              header = excluded.header,
              body = excluded.body,
              update_time = excluded.update_time,
              keywords = excluded.keywords,
              description = excluded.description,
              status = excluded.status,
              cover = excluded.cover,
              owner = excluded.owner,
              channel = excluded.channel,
              partition = excluded.partition,
              version = excluded.version,
              build = excluded.build,
              url = excluded.url,
              branch = excluded.branch,
              commit = excluded.commit,
              commit_time = excluded.commit_time,
              relative_path = excluded.relative_path,
              repo_id = excluded.repo_id,
              lang = excluded.lang,
              name = excluded.name,
              checksum = excluded.checksum,
              syncno = excluded.syncno,
			  repo_first_commit=excluded.repo_first_commit;`

	paramsMap := map[string]interface{}{
		"uid":               noteTable.Uid,
		"title":             noteTable.Title,
		"header":            noteTable.Header,
		"body":              noteTable.Body,
		"create_time":       noteTable.CreateTime,
		"update_time":       noteTable.UpdateTime,
		"keywords":          noteTable.Keywords,
		"description":       noteTable.Description,
		"status":            noteTable.Status,
		"cover":             noteTable.Cover,
		"owner":             noteTable.Owner,
		"channel":           noteTable.Channel,
		"discover":          noteTable.Discover,
		"partition":         noteTable.Partition,
		"version":           noteTable.Version,
		"build":             noteTable.Build,
		"url":               noteTable.Url,
		"branch":            noteTable.Branch,
		"commit":            noteTable.Commit,
		"commit_time":       noteTable.CommitTime,
		"relative_path":     noteTable.RelativePath,
		"repo_id":           noteTable.RepoId,
		"lang":              noteTable.Lang,
		"name":              noteTable.Name,
		"checksum":          noteTable.Checksum,
		"syncno":            noteTable.Syncno,
		"repo_first_commit": noteTable.RepoFirstCommit,
	}

	_, err := datastore.NamedExec(sqlText, paramsMap)
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
	pagination.Count = countSqlResults[0].Count

	//selectData := &nemodels.NESelectResult[*MTNoteModel]{
	//	Page:  pagination.Page,
	//	Size:  pagination.Size,
	//	Count: countSqlResults[0].Count,
	//	Range: resultRange,
	//}

	return pagination, sqlResults, nil
}

func PGGetNoteByChecksum(checksum string) (*MTNoteTable, error) {
	if checksum == "" {
		return nil, fmt.Errorf("PGGetNote uid is empty")
	}
	pageSqlText := ` select * from articles where checksum= :checksum limit 1; `

	pageSqlParams := map[string]interface{}{
		"checksum": checksum,
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

// PGGetNote 获取单个笔记信息
func PGGetNote(uid string, lang string) (*MTNoteTable, error) {
	if uid == "" {
		return nil, fmt.Errorf("PGGetNote uid is empty")
	}
	pageSqlText := ` select * from articles where status = 1 and uid = :uid; `

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
