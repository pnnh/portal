package notes

import (
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"portal/models"
	"portal/neutron/helpers"
	"portal/neutron/services/datastore"
	"time"
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
	Owner       sql.NullString `json:"-"`
	Channel     sql.NullString `json:"-"`
	Discover    int            `json:"discover"`
	Partition   sql.NullString `json:"-"`
	CreateTime  time.Time      `json:"create_time" db:"create_time"`
	UpdateTime  time.Time      `json:"update_time" db:"update_time"`
}

func PGInsertNote(model *MTNoteModel) error {
	sqlText := `insert into articles(uid, title, header, body, description, create_time, update_time)
values(:uid, :title, :header, :body, :description, now(), now())
on conflict (uid)
do update set title=excluded.title, body=excluded.body, update_time = now();`

	sqlParams := map[string]interface{}{
		"uid":         model.Uid,
		"title":       model.Title,
		"header":      "MTNote",
		"body":        model.Body,
		"description": model.Description,
	}

	_, err := datastore.NamedExec(sqlText, sqlParams)
	if err != nil {
		return fmt.Errorf("PGInsertNote: %w", err)
	}
	return nil
}

func SelectNotes(page int, size int) (*models.SelectData, error) {

	pagination := helpers.CalcPaginationByPage(page, size)
	baseSqlText := ` select * from articles where status = 1 order by create_time desc `

	pageSqlText := baseSqlText + ` offset :offset limit :limit; `
	pageSqlParams := map[string]interface{}{
		"offset": pagination.Offset, "limit": pagination.Limit}
	var sqlResults []*MTNoteModel

	rows, err := datastore.NamedQuery(pageSqlText, pageSqlParams)
	if err != nil {
		return nil, fmt.Errorf("NamedQuery: %w", err)
	}
	if err = sqlx.StructScan(rows, &sqlResults); err != nil {
		return nil, fmt.Errorf("StructScan: %w", err)
	}

	resultRange := make([]any, 0)
	for _, item := range sqlResults {
		resultRange = append(resultRange, item)
	}

	countSqlText := `select count(1) as count from (` + baseSqlText + `) as temp;`

	countSqlParams := map[string]interface{}{}
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

	selectData := &models.SelectData{
		Page:  pagination.Page,
		Size:  pagination.Size,
		Count: countSqlResults[0].Count,
		Range: resultRange,
	}

	return selectData, nil
}