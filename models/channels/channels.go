package channels

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"neutron/helpers"
	"neutron/services/datastore"
	"portal/models"
)

type MTChannelModel struct {
	Uid         string         `json:"uid"`
	Name        string         `json:"name"`
	Description sql.NullString `json:"description"`
	Image       sql.NullString `json:"image"`
	Status      int            `json:"status"`
	CreateTime  time.Time      `json:"create_time" db:"create_time"`
	UpdateTime  time.Time      `json:"update_time" db:"update_time"`
	Cid         sql.NullString `json:"cid" db:"cid"`
	Lang        sql.NullString `json:"lang" db:"lang"`
	Owner       string         `json:"owner" db:"owner"`
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
	if m.Cid.Valid {
		view.Cid = m.Cid.String
	}
	if m.Lang.Valid {
		view.Lang = m.Lang.String
	}
	return view
}

type MTChannelView struct {
	MTChannelModel
	Description string `json:"description"`
	Image       string `json:"image"`
	Cid         string `json:"cid"`
	Lang        string `json:"lang"`
	Match       string `json:"match"` // 用于自动完成时的匹配
}

func SelectChannels(keyword string, page int, size int, lang string) (*models.SelectResult[MTChannelModel], error) {
	pagination := helpers.CalcPaginationByPage(page, size)
	baseSqlText := ` select * from channels `
	baseSqlParams := map[string]interface{}{}

	whereText := ` where status = 1 `
	if keyword != "" {
		whereText += ` and (name like :keyword or description like :keyword) `
		baseSqlParams["keyword"] = "%" + keyword + "%"
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

// 输入频道时自动完成
func CompleteChannels(keyword string, lang string) (*models.SelectResult[MTChannelView], error) {
	if keyword == "" {
		return nil, fmt.Errorf("keyword cannot be empty")
	}
	baseSqlText := ` select * from channels `
	baseSqlParams := map[string]interface{}{}

	whereText := ` where status = 1 `
	exactMatch := false
	if helpers.IsUuid(keyword) {
		exactMatch = true
		whereText += ` and uid = :keyword or (cid = :keyword and lang = :lang) `
		baseSqlParams["lang"] = lang
		baseSqlParams["keyword"] = keyword
	} else {
		whereText += ` and (name ilike :keyword or description ilike :keyword) `
		baseSqlParams["keyword"] = "%" + keyword + "%"
	}
	orderText := ` order by create_time desc `

	pageSqlText := fmt.Sprintf("%s %s %s %s", baseSqlText, whereText, orderText, ` limit :limit; `)
	limitSize := 10 // 限制返回10条数据
	pageSqlParams := map[string]interface{}{
		"limit": limitSize,
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

	resultRange := make([]MTChannelView, 0)
	for _, item := range sqlResults {
		viewModel := item.ToViewModel().(*MTChannelView)
		if exactMatch {
			viewModel.Match = "exact"
		}
		resultRange = append(resultRange, *viewModel)
	}
	if len(resultRange) == 1 && strings.ToLower(resultRange[0].Name) == strings.ToLower(keyword) {
		resultRange[0].Match = "exact"
	}

	selectData := &models.SelectResult[MTChannelView]{
		Page:  1,
		Size:  limitSize,
		Count: len(resultRange),
		Range: resultRange,
	}

	return selectData, nil
}

func PGGetChannel(uid, lang string) (*MTChannelModel, error) {
	pageSqlText := ` select * from channels where status = 1 and  ((uid = :uid) or (cid = :uid and lang = :lang)); `
	pageSqlParams := map[string]interface{}{
		"uid":  uid,
		"lang": lang,
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
