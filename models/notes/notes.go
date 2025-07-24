package notes

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"neutron/helpers"
	"neutron/services/datastore"
	"portal/models"
)

type MTNoteMatter struct {
	Cls         string `json:"cls"`
	Uid         string `json:"uid"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type MTNoteModel struct {
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
	Nid         int64          `json:"nid" db:"nid"`
	Dc          string         `json:"-" db:"dc"`
}

func (m MTNoteModel) ToViewModel() interface{} {
	view := &MTNoteView{
		MTNoteModel: m,
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

type MTNoteView struct {
	MTNoteModel
	Cid  string `json:"cid" db:"cid"`
	Lang string `json:"lang" db:"lang"`
}

func PGInsertNote(model *MTNoteModel) error {
	sqlText := `insert into articles(uid, title, header, body, description, create_time, update_time, 
                     version, build, url, branch, commit, commit_time, repo_path, repo_id, cid, lang, dc, channel, owner)
values(:uid, :title, :header, :body, :description, now(), now(), :version, :build, :url, :branch, 
       :commit, :commit_time, :repo_path, :repo_id, :cid, :lang, :dc, :channel, :owner)
on conflict (uid)
do nothing;`

	sqlParams := map[string]interface{}{
		"uid":         model.Uid,
		"title":       model.Title,
		"header":      "MTNote",
		"body":        model.Body,
		"description": model.Description,
		"version":     model.Version,
		"build":       model.Build,
		"url":         model.Url,
		"branch":      model.Branch,
		"commit":      model.Commit,
		"commit_time": model.CommitTime,
		"repo_path":   model.RepoPath,
		"repo_id":     model.RepoId,
		"cid":         model.Cid,
		"lang":        model.Lang,
		"dc":          model.Dc,
		"channel":     model.Channel,
		"owner":       model.Owner,
	}

	_, err := datastore.NamedExec(sqlText, sqlParams)
	if err != nil {
		return fmt.Errorf("PGInsertNote: %w", err)
	}
	return nil
}

func PGUpdateNote(model *MTNoteModel) error {
	sqlText := `update articles set title = :title,  body = :body, description = :description, 
	update_time = now() where uid = :uid;`

	sqlParams := map[string]interface{}{
		"uid":         model.Uid,
		"title":       model.Title,
		"body":        model.Body,
		"description": model.Description,
	}

	if _, err := datastore.NamedExec(sqlText, sqlParams); err != nil {
		return fmt.Errorf("PGUpdateNote: %w", err)
	}
	return nil
}

func SelectNotes(channel, keyword string, page int, size int, lang string) (*models.SelectResult[MTNoteModel], error) {
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
	var sqlResults []MTNoteModel

	rows, err := datastore.NamedQuery(pageSqlText, pageSqlParams)
	if err != nil {
		return nil, fmt.Errorf("NamedQuery: %w", err)
	}
	if err = sqlx.StructScan(rows, &sqlResults); err != nil {
		return nil, fmt.Errorf("StructScan: %w", err)
	}

	resultRange := make([]MTNoteModel, 0)
	for _, item := range sqlResults {
		resultRange = append(resultRange, item)
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
		return nil, fmt.Errorf("NamedQuery: %w", err)
	}
	if err = sqlx.StructScan(rows, &countSqlResults); err != nil {
		return nil, fmt.Errorf("StructScan: %w", err)
	}
	if len(countSqlResults) == 0 {
		return nil, fmt.Errorf("查询笔记总数有误，数据为空")
	}

	selectData := &models.SelectResult[MTNoteModel]{
		Page:  pagination.Page,
		Size:  pagination.Size,
		Count: countSqlResults[0].Count,
		Range: resultRange,
	}

	return selectData, nil
}

// PGGetNote 获取单个笔记信息
// obsolete 处理向后兼容逻辑。早期是通过单个uid查询，后期可以通过cid和lang查询。
func PGGetNote(uid string, lang string) (*MTNoteModel, error) {
	if uid == "" {
		return nil, fmt.Errorf("PGGetNote uid is empty")
	}
	pageSqlText := ` select * from articles where status = 1 and (uid = :uid or (cid = :uid and lang = :lang)); `

	pageSqlParams := map[string]interface{}{
		"uid":  uid,
		"lang": lang,
	}
	var sqlResults []*MTNoteModel

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
