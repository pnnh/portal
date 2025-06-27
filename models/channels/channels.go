package channels

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"portal/models"
	"portal/neutron/helpers"
	"portal/neutron/services/datastore"
)

type MTChannelModel struct {
	Uid         string         `json:"uid"`
	Name        string         `json:"name"`
	Description sql.NullString `json:"description"`
	Image       sql.NullString `json:"image"`
	Status      int            `json:"status"`
	CreateTime  time.Time      `json:"create_time" db:"create_time"`
	UpdateTime  time.Time      `json:"update_time" db:"update_time"`
}

func (m MTChannelModel) ToViewModel() interface{} {
	view := &MTChannelView{
		MTChannelModel: m,
	}
	if m.Image.Valid {
		view.Image = m.Image.String
	}
	if m.Description.Valid {
		view.Description = m.Description.String
	}
	return view
}

type MTChannelView struct {
	MTChannelModel
	Description string `json:"description"`
	Image       string `json:"image"`
}

func SelectChannels(keyword string, page int, size int) (*models.SelectResult[MTChannelModel], error) {
	pagination := helpers.CalcPaginationByPage(page, size)
	baseSqlText := ` select * from channels `
	baseSqlParams := map[string]interface{}{}

	whereText := ` where status = 1 `
	if keyword != "" {
		whereText += ` and (name like :keyword or description like :keyword) `
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
	var sqlResults []*MTChannelModel

	rows, err := datastore.NamedQuery(pageSqlText, pageSqlParams)
	if err != nil {
		return nil, fmt.Errorf("NamedQuery: %w", err)
	}
	if err = sqlx.StructScan(rows, &sqlResults); err != nil {
		return nil, fmt.Errorf("StructScan: %w", err)
	}

	resultRange := make([]MTChannelModel, 0)
	for _, item := range sqlResults {
		resultRange = append(resultRange, *item)
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

	selectData := &models.SelectResult[MTChannelModel]{
		Page:  pagination.Page,
		Size:  pagination.Size,
		Count: countSqlResults[0].Count,
		Range: resultRange,
	}

	return selectData, nil
}

func PGGetChannel(uid string) (*MTChannelModel, error) {
	pageSqlText := ` select * from channels where status = 1 and uid = :uid; `
	pageSqlParams := map[string]interface{}{
		"uid": uid,
	}
	var sqlResults []*MTChannelModel

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
