package images

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"portal/models"
	"portal/neutron/helpers"
	"portal/neutron/services/datastore"
)

type MTImageModel struct {
	Uid         string         `json:"uid"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Keywords    string         `json:"keywords"`
	Status      int            `json:"status"`
	Owner       sql.NullString `json:"-"`
	Channel     sql.NullString `json:"-"`
	Discover    int            `json:"discover"`
	CreateTime  time.Time      `json:"create_time" db:"create_time"`
	UpdateTime  time.Time      `json:"update_time" db:"update_time"`
	FilePath    string         `json:"file_path" db:"file_path"`
	ExtName     string         `json:"ext_name" db:"ext_name"`
}

func PGInsertImage(model *MTImageModel) error {
	sqlText := `insert into images(uid, title, keywords, description, create_time, update_time, channel, status, 
                   discover, owner, file_path, ext_name)
values(:uid, :title, :keywords, :description, now(), now(), :channel, :status, :discover, :owner, :file_path, :ext_name)
on conflict (uid)
do update set title=excluded.title, description=excluded.description, update_time = now(),
	keywords=excluded.keywords, channel=excluded.channel, file_path=excluded.file_path, ext_name=excluded.ext_name;`

	sqlParams := map[string]interface{}{
		"uid":         model.Uid,
		"title":       model.Title,
		"description": model.Description,
		"keywords":    model.Keywords,
		"channel":     model.Channel,
		"status":      model.Status,
		"discover":    model.Discover,
		"owner":       model.Owner,
		"file_path":   model.FilePath,
		"ext_name":    model.ExtName,
	}

	_, err := datastore.NamedExec(sqlText, sqlParams)
	if err != nil {
		return fmt.Errorf("PGInsertImage: %w", err)
	}
	return nil
}

func SelectImages(keyword string, page int, size int) (*models.SelectData, error) {
	pagination := helpers.CalcPaginationByPage(page, size)
	baseSqlText := ` select * from images `
	baseSqlParams := map[string]interface{}{}

	whereText := ` where (status = 1 or discover < 10) `
	if keyword != "" {
		whereText += ` and (title like :keyword or description like :keyword) `
		baseSqlParams["keyword"] = "%" + keyword + "%"
	}
	orderText := ` order by create_time desc `

	pageSqlText := fmt.Sprintf("%s %s %s %s", baseSqlText, whereText, orderText, ` offset :offset limit :limit; `)
	pageSqlParams := map[string]interface{}{
		"offset": pagination.Offset, "limit": pagination.Limit,
	}
	for k, v := range baseSqlParams {
		pageSqlParams[k] = v
	}
	var sqlResults []*MTImageModel

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
		return nil, fmt.Errorf("查询图片总数有误，数据为空")
	}

	selectData := &models.SelectData{
		Page:  pagination.Page,
		Size:  pagination.Size,
		Count: countSqlResults[0].Count,
		Range: resultRange,
	}

	return selectData, nil
}

func PGGetImage(uid string) (*MTImageModel, error) {

	pageSqlText := ` select * from images where (status = 1 or discover < 10) and uid = :uid; `
	pageSqlParams := map[string]interface{}{
		"uid": uid,
	}
	var sqlResults []*MTImageModel

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
