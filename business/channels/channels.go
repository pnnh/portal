package channels

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	nemodels "neutron/models"

	"neutron/helpers"
	"neutron/services/datastore"

	"github.com/jmoiron/sqlx"
)

type MTChannelTable struct {
	Error       error          `json:"-" db:"-"`
	Uid         string         `json:"uid"`
	Name        string         `json:"name"`
	Title       sql.NullString `json:"title"`
	Description sql.NullString `json:"description"`
	Image       sql.NullString `json:"image"`
	Status      int            `json:"status"`
	CreateTime  time.Time      `json:"create_time" db:"create_time"`
	UpdateTime  time.Time      `json:"update_time" db:"update_time"`
	Lang        sql.NullString `json:"lang" db:"lang"`
	Owner       string         `json:"owner" db:"owner"`
}

func (m *MTChannelTable) FromMap(tableMap *datastore.DataRow) *MTChannelTable {
	if tableMap == nil {
		m.Error = fmt.Errorf("tableMap cannot be nil")
		return m
	}
	m.Uid = tableMap.GetString("uid")
	m.Name = tableMap.GetString("name")
	m.Title = tableMap.GetNullString("title")
	m.Description = tableMap.GetNullString("description")
	m.Image = tableMap.GetNullString("image")
	m.Status = tableMap.GetInt("status")
	m.CreateTime = tableMap.GetTime("create_time")
	m.UpdateTime = tableMap.GetTime("update_time")
	m.Lang = tableMap.GetNullString("lang")
	m.Owner = tableMap.GetString("owner")

	return m
}

func (m *MTChannelTable) TableName() string {
	return "channels"
}

func (m *MTChannelTable) PGGetByUid(uid string) *MTChannelTable {
	getMap, err := datastore.NewGetQuery("channels",
		"status = 1 and uid = :uid", "", "",
		map[string]any{"uid": uid})
	if err != nil {
		m.Error = fmt.Errorf("NewGetQuery: %w", err)
		return m
	}
	table := m.FromMap(getMap)
	if table.Error != nil {
		m.Error = fmt.Errorf("FromTableMap: %w", err)
		return m
	}
	return table
}

func (m *MTChannelTable) ToModel() *MTChannelModel {
	return &MTChannelModel{
		MTChannelTable: *m,
		Title:          m.Title.String,
		Description:    m.Description.String,
		Image:          m.Image.String,
		Lang:           m.Lang.String,
	}
}

type MTChannelModel struct {
	MTChannelTable
	Title       string `json:"title"`
	Description string `json:"description"`
	Image       string `json:"image"`
	Lang        string `json:"lang"`
}

func (m *MTChannelModel) FromTable(table *MTChannelTable) (*MTChannelModel, error) {
	if table == nil {
		return nil, fmt.Errorf("table cannot be nil")
	}
	m.Uid = table.Uid
	m.Name = table.Name
	m.Title = table.Title.String
	m.Description = table.Description.String
	m.Image = table.Image.String
	m.Status = table.Status
	m.CreateTime = table.CreateTime
	m.UpdateTime = table.UpdateTime
	m.Lang = table.Lang.String
	m.Owner = table.Owner

	return m, nil
}

//func (m *MTChannelModel) GetByUid(uid string) (*MTChannelModel, error) {
//	table:= (&MTChannelTable{}).PGGetByUid(uid)
//	if table.Error != nil {
//		return nil, fmt.Errorf("FromTableMap: %w", table.Error)
//	}
//	return m.FromTable(table)
//}

func (m MTChannelModel) ToViewModel() interface{} {
	view := &MTChannelView{
		MTChannelModel: m,
	}

	return view
}

type MTChannelView struct {
	MTChannelModel
	Match string `json:"match"` // 用于自动完成时的匹配
}

func SelectChannels(keyword string, page int, size int, lang string) (*nemodels.NESelectResult[MTChannelModel], error) {
	pagination := helpers.CalcPaginationByPage(page, size)
	baseSqlText := ` select * from channels `
	baseSqlParams := map[string]interface{}{}

	whereText := ` where status = 1 `
	if keyword != "" {
		whereText += ` and (name like :keyword or description like :keyword) `
		baseSqlParams["keyword"] = "%" + keyword + "%"
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

	selectData := &nemodels.NESelectResult[MTChannelModel]{
		Page:  pagination.Page,
		Size:  pagination.Size,
		Count: countSqlResults[0].Count,
		Range: resultRange,
	}

	return selectData, nil
}

// 输入频道时自动完成
func PGCompleteChannels(keyword string, lang string) (*nemodels.NESelectResult[MTChannelView], error) {
	if keyword == "" {
		return nil, fmt.Errorf("keyword cannot be empty")
	}
	baseSqlText := ` select * from channels `
	baseSqlParams := map[string]interface{}{}

	whereText := ` where status = 1 `
	exactMatch := false
	if helpers.IsUuid(keyword) {
		exactMatch = true
		whereText += ` and uid = :keyword `
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

	selectData := &nemodels.NESelectResult[MTChannelView]{
		Page:  1,
		Size:  limitSize,
		Count: len(resultRange),
		Range: resultRange,
	}

	return selectData, nil
}

func PGGetChannelByCid(cid, lang string) (*MTChannelModel, error) {
	pageSqlText := ` select * from channels where status = 1 and cid = :cid and lang = :lang; `
	pageSqlParams := map[string]interface{}{
		"cid":  cid,
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

func PGConsoleGetChannelByUid(uid string) (*MTChannelModel, error) {
	pageSqlText := ` select * from channels where uid = :uid; `
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
