package channels

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	nemodels "neutron/models"

	"neutron/helpers"
	"neutron/services/datastore"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type MTChantransTable struct {
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

func (m *MTChantransTable) FromMap(tableMap *datastore.DataRow) *MTChantransTable {
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

func (m *MTChantransTable) TableName() string {
	return "channels"
}

func (m *MTChantransTable) PGGetByUid(uid string) *MTChantransTable {
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

func (m *MTChantransTable) ToModel() *MTChantransModel {
	return &MTChantransModel{
		MTChantransTable: *m,
		Title:            m.Title.String,
		Description:      m.Description.String,
		Image:            m.Image.String,
		Lang:             m.Lang.String,
	}
}

type MTChantransModel struct {
	MTChantransTable
	Title       string `json:"title"`
	Description string `json:"description"`
	Image       string `json:"image"`
	Lang        string `json:"lang"`
}

func (m *MTChantransModel) FromTable(table *MTChantransTable) (*MTChantransModel, error) {
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

//func (m *MTChantransModel) GetByUid(uid string) (*MTChantransModel, error) {
//	table:= (&MTChantransTable{}).PGGetByUid(uid)
//	if table.Error != nil {
//		return nil, fmt.Errorf("FromTableMap: %w", table.Error)
//	}
//	return m.FromTable(table)
//}

func (m MTChantransModel) ToViewModel() interface{} {
	view := &MTChantransView{
		MTChantransModel: m,
	}

	return view
}

type MTChantransView struct {
	MTChantransModel
	Match string `json:"match"` // 用于自动完成时的匹配
}

func SelectChantranss(keyword string, page int, size int, lang string) (*nemodels.NESelectResult[MTChantransModel], error) {
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
	var sqlResults []*MTChantransModel

	rows, err := datastore.NamedQuery(pageSqlText, pageSqlParams)
	if err != nil {
		return nil, fmt.Errorf("NamedQuery: %w", err)
	}
	if err = sqlx.StructScan(rows, &sqlResults); err != nil {
		return nil, fmt.Errorf("StructScan: %w", err)
	}

	resultRange := make([]MTChantransModel, 0)
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

	selectData := &nemodels.NESelectResult[MTChantransModel]{
		Page:  pagination.Page,
		Size:  pagination.Size,
		Count: countSqlResults[0].Count,
		Range: resultRange,
	}

	return selectData, nil
}

// 输入频道时自动完成
func PGCompleteChantranss(keyword string, lang string) (*nemodels.NESelectResult[MTChantransView], error) {
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
	var sqlResults []*MTChantransModel

	rows, err := datastore.NamedQuery(pageSqlText, pageSqlParams)
	if err != nil {
		return nil, fmt.Errorf("NamedQuery: %w", err)
	}
	if err = sqlx.StructScan(rows, &sqlResults); err != nil {
		return nil, fmt.Errorf("StructScan: %w", err)
	}

	resultRange := make([]MTChantransView, 0)
	for _, item := range sqlResults {
		viewModel := item.ToViewModel().(*MTChantransView)
		if exactMatch {
			viewModel.Match = "exact"
		}
		resultRange = append(resultRange, *viewModel)
	}
	if len(resultRange) == 1 && strings.ToLower(resultRange[0].Name) == strings.ToLower(keyword) {
		resultRange[0].Match = "exact"
	}

	selectData := &nemodels.NESelectResult[MTChantransView]{
		Page:  1,
		Size:  limitSize,
		Count: len(resultRange),
		Range: resultRange,
	}

	return selectData, nil
}

func PGGetChantransByCid(cid, lang string) (*MTChantransModel, error) {
	pageSqlText := ` select * from channels where status = 1 and cid = :cid and lang = :lang; `
	pageSqlParams := map[string]interface{}{
		"cid":  cid,
		"lang": lang,
	}
	var sqlResults []*MTChantransModel

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

func PGConsoleGetChantransByUid(uid string) (*MTChantransModel, error) {
	pageSqlText := ` select * from channels where uid = :uid; `
	pageSqlParams := map[string]interface{}{
		"uid": uid,
	}
	var sqlResults []*MTChantransModel

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

func ChantransSelectHandler(gctx *gin.Context) {
	keyword := gctx.Query("keyword")
	page := gctx.Query("page")
	size := gctx.Query("size")
	lang := gctx.Query("lang")
	pageInt, err := strconv.Atoi(page)
	if err != nil {
		pageInt = 1
	}
	sizeInt, err := strconv.Atoi(size)
	if err != nil {
		sizeInt = 10
	}
	if lang == "" {
		lang = nemodels.DefaultLanguage
	}
	selectResult, err := SelectChantranss(keyword, pageInt, sizeInt, lang)
	if err != nil {
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "查询频道出错"))
		return
	}

	selectResponse := nemodels.NESelectResultToResponse(selectResult)
	responseResult := nemodels.NECodeOk.WithData(selectResponse)

	gctx.JSON(http.StatusOK, responseResult)
}

// 输入时自动完成
func ChantransCompleteHandler(gctx *gin.Context) {
	keyword := gctx.Query("keyword")
	lang := gctx.Query("lang")
	selectResult, err := PGCompleteChantranss(keyword, lang)
	if err != nil {
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "查询频道出错"))
		return
	}

	responseResult := nemodels.NECodeOk.WithData(selectResult)

	gctx.JSON(http.StatusOK, responseResult)
}

func ChantransGetByCidHandler(gctx *gin.Context) {
	cid := gctx.Param("cid")
	lang := gctx.Param("lang")
	wangLang := gctx.Param("wantLang")
	if lang == "" {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithLocalMessage(nemodels.LangEn,
			"lang不能为空", "lang cannot be empty"))
		return
	}
	if cid == "" || wangLang == "" {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithLocalMessage(lang,
			"cid或wantLang不能为空", "cid or wantLang cannot be empty"))
		return
	}

	selectResult, err := PGGetChantransByCid(cid, wangLang)
	if err != nil {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithLocalError(lang, err, "查询频道出错", "query channel failed"))
		return
	}
	if selectResult == nil {
		gctx.JSON(http.StatusOK, nemodels.NECodeNotFound.WithLocalMessage(lang, "频道不存在", "channel not found"))
		return
	}
	var modelData = selectResult.ToViewModel()

	responseResult := nemodels.NECodeOk.WithData(modelData)

	gctx.JSON(http.StatusOK, responseResult)
}
