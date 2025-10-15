package notes

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	nemodels "neutron/models"
	"neutron/services/datetime"

	"neutron/helpers"
	"neutron/services/datastore"
	"portal/business"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

func ConsoleNotesSelectHandler(gctx *gin.Context) {
	keyword := gctx.Query("keyword")
	page := gctx.Query("page")
	size := gctx.Query("size")
	channel := gctx.Query("channel")
	lang := gctx.Query("lang")
	pageInt, err := strconv.Atoi(page)
	if err != nil {
		pageInt = 1
	}
	sizeInt, err := strconv.Atoi(size)
	if err != nil {
		sizeInt = 10
	}
	accountModel, err := business.FindAccountFromCookie(gctx)
	if err != nil {
		logrus.Warnln("ConsoleNotesSelectHandler", err)
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "查询账号出错b"))
		return
	}
	if accountModel == nil {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("账号不存在"))
		return
	}
	pagination, selectResult, err := ConsoleSelectNotes(accountModel.Uid, channel, keyword, pageInt, sizeInt, lang)
	if err != nil {
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "查询笔记出错"))
		return
	}
	respView := make([]map[string]interface{}, 0)
	for _, v := range selectResult {
		outView, err := consoleNoteGetOutView(v)
		if err != nil {
			gctx.JSON(http.StatusOK, nemodels.NECodeError.WithError(err))
			return
		}
		respView = append(respView, outView)
	}

	resp := map[string]any{
		"page":  pagination.Page,
		"size":  pagination.Size,
		"count": pagination.Count,
		"range": respView,
	}

	responseResult := nemodels.NECodeOk.WithData(resp)

	gctx.JSON(http.StatusOK, responseResult)
}

func ConsoleSelectNotes(owner, channel, keyword string, page int, size int, lang string) (*helpers.Pagination,
	[]*datastore.DataRow, error) {
	pagination := helpers.CalcPaginationByPage(page, size)
	baseSqlText := ` select a.*, c.name channel_name from articles as a left join channels as c on a.channel = c.uid `
	baseSqlParams := map[string]interface{}{}

	whereText := ` where a.owner = :owner `
	baseSqlParams["owner"] = owner
	if keyword != "" {
		whereText += ` and (a.title like :keyword or a.description like :keyword) `
		baseSqlParams["keyword"] = "%" + keyword + "%"
	}
	if channel != "" {
		whereText += ` and a.channel = :channel `
		baseSqlParams["channel"] = channel
	}
	if lang != "" {
		whereText += ` and a.lang = :lang `
		baseSqlParams["lang"] = lang
	}
	orderText := ` order by a.create_time desc `

	pageSqlText := fmt.Sprintf("%s %s %s %s", baseSqlText, whereText, orderText, ` offset :offset limit :limit; `)
	pageSqlParams := map[string]interface{}{
		"offset": pagination.Offset, "limit": pagination.Limit,
	}
	for k, v := range baseSqlParams {
		pageSqlParams[k] = v
	}
	var sqlResults = make([]*datastore.DataRow, 0)

	rows, err := datastore.NamedQuery(pageSqlText, pageSqlParams)
	if err != nil {
		return nil, nil, fmt.Errorf("NewSelectQuery: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			logrus.Warnf("rows.Close: %w", closeErr)
		}
	}()

	for rows.Next() {
		rowMap := make(map[string]interface{})
		if err := rows.MapScan(rowMap); err != nil {
			return nil, nil, fmt.Errorf("MapScan: %w", err)
		}
		tableMap := datastore.MapToDataRow(rowMap)
		sqlResults = append(sqlResults, tableMap)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("rows error: %w", err)
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
		return nil, sqlResults, fmt.Errorf("NamedQuery: %w", err)
	}
	if err = sqlx.StructScan(rows, &countSqlResults); err != nil {
		return nil, sqlResults, fmt.Errorf("StructScan: %w", err)
	}
	if len(countSqlResults) == 0 {
		return nil, sqlResults, fmt.Errorf("查询笔记总数有误，数据为空")
	}
	pagination.Count = countSqlResults[0].Count
	if pagination.Count == 0 {
		return pagination, sqlResults, nil
	}

	return pagination, sqlResults, nil
}

func consoleNoteGetOutView(dataRow *datastore.DataRow) (map[string]interface{}, error) {
	outView := make(map[string]interface{})
	outView["uid"] = dataRow.GetString("uid")
	outView["title"] = dataRow.GetString("title")
	outView["header"] = dataRow.GetString("header")
	outView["body"] = dataRow.GetString("body")
	outView["description"] = dataRow.GetString("description")
	outView["keywords"] = dataRow.GetString("keywords")
	outView["status"] = dataRow.GetInt("status")
	outView["cover"] = dataRow.GetStringOrDefault("cover", "")
	outView["owner"] = dataRow.GetString("owner")
	outView["channel"] = dataRow.GetStringOrDefault("channel", "")
	outView["discover"] = dataRow.GetInt("discover")
	outView["partition"] = dataRow.GetStringOrDefault("partition", "")
	outView["create_time"] = dataRow.GetTime("create_time")
	outView["update_time"] = dataRow.GetTime("update_time")
	outView["version"] = dataRow.GetStringOrDefault("version", "")
	outView["build"] = dataRow.GetStringOrDefault("build", "")
	outView["url"] = dataRow.GetStringOrDefault("url", "")
	outView["branch"] = dataRow.GetStringOrDefault("branch", "")
	outView["commit"] = dataRow.GetStringOrDefault("commit", "")
	outView["commit_time"] = dataRow.GetTimeOrDefault("commit_time", datetime.UtcMinTime)
	outView["relative_path"] = dataRow.GetStringOrDefault("relative_path", "")
	outView["repo_id"] = dataRow.GetStringOrDefault("repo_id", "")
	outView["lang"] = dataRow.GetStringOrDefault("lang", "")
	outView["name"] = dataRow.GetStringOrDefault("name", "")
	outView["checksum"] = dataRow.GetStringOrDefault("checksum", "")
	outView["syncno"] = dataRow.GetStringOrDefault("syncno", "")
	outView["repo_first_commit"] = dataRow.GetStringOrDefault("repo_first_commit", "")
	outView["channel_name"] = dataRow.GetStringOrDefault("channel_name", "")

	return outView, nil
}

func ConsoleNoteGetHandler(gctx *gin.Context) {
	uid := gctx.Param("uid")
	wantLang := gctx.Query("wantLang")
	if uid == "" {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("uid不能为空"))
		return
	}
	accountModel, err := business.FindAccountFromCookie(gctx)
	if err != nil {
		logrus.Warnln("ConsoleNotesSelectHandler", err)
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "查询账号出错b"))
		return
	}
	if accountModel == nil {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("账号不存在"))
		return
	}
	selectResult, err := PGConsoleGetNote(accountModel.Uid, uid, wantLang)
	if err != nil {
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "查询笔记出错"))
		return
	}
	if selectResult == nil {
		gctx.JSON(http.StatusOK, nemodels.NECodeNotFound)
		return
	}
	model := selectResult.ToModel()
	responseResult := nemodels.NECodeOk.WithData(model)

	gctx.JSON(http.StatusOK, responseResult)
}

func ConsoleNoteDeleteHandler(gctx *gin.Context) {
	uid := gctx.Param("uid")
	lang := gctx.Query("lang")
	if uid == "" {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("uid不能为空"))
		return
	}
	accountModel, err := business.FindAccountFromCookie(gctx)
	if err != nil {
		logrus.Warnln("ConsoleNotesSelectHandler", err)
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "查询账号出错b"))
		return
	}
	if accountModel == nil {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("账号不存在"))
		return
	}
	err = PGConsoleDeleteNote(accountModel.Uid, uid, lang)
	if err != nil {
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "查询笔记出错"))
		return
	}
	responseResult := nemodels.NECodeOk.WithData(uid)

	gctx.JSON(http.StatusOK, responseResult)
}

func PGConsoleGetNote(owner, uid string, wantLang string) (*MTNoteTable, error) {
	if uid == "" {
		return nil, fmt.Errorf("PGConsoleGetNote uid is empty")
	}

	pageSqlParams := map[string]interface{}{
		"owner": owner,
	}
	var pageSqlText string
	if business.IsSupportedLanguage(wantLang) {
		pageSqlText = ` select * from articles where owner = :owner and (cid = :cid and lang = :wantLang); `
		pageSqlParams["cid"] = uid
		pageSqlParams["wantLang"] = wantLang
	} else {
		pageSqlText = ` select * from articles where owner = :owner and uid = :uid; `
		pageSqlParams["uid"] = uid
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

func PGConsoleDeleteNote(owner, uid string, lang string) error {
	if uid == "" {
		return fmt.Errorf("PGConsoleGetNote uid is empty")
	}
	pageSqlText := ` delete from articles where (owner = :owner and (uid = :uid or (cid = :uid and lang = :lang))); `

	pageSqlParams := map[string]interface{}{
		"uid":   uid,
		"lang":  lang,
		"owner": owner,
	}

	_, err := datastore.NamedQuery(pageSqlText, pageSqlParams)
	if err != nil {
		return fmt.Errorf("NamedQuery: %w", err)
	}

	return nil
}

func ConsoleNoteUpdateHandler(gctx *gin.Context) {
	uid := gctx.Param("uid")
	if uid == "" {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("uid不能为空"))
		return
	}
	accountModel, err := business.FindAccountFromCookie(gctx)
	if err != nil {
		logrus.Warnln("NoteUpdateHandler", err)
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "查询账号出错c"))
		return
	}
	if accountModel == nil || accountModel.IsAnonymous() {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("账号不存在或匿名用户不能修改笔记"))
		return
	}

	model := &MTNoteModel{}
	if err := gctx.ShouldBindJSON(model); err != nil {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithError(err))
		return
	}
	if model.Title == "" || model.Body == "" {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("标题或内容不能为空"))
		return
	}
	oldModel, err := PGConsoleGetNote(accountModel.Uid, uid, "")
	if err != nil {
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "查询笔记出错"))
		return
	}
	if oldModel == nil {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("笔记不存在"))
		return
	}
	if oldModel.Owner != accountModel.Uid {
		gctx.JSON(http.StatusOK, nemodels.NECodeUnauthorized.WithMessage("没有权限修改该笔记"))
		return
	}

	model.Uid = uid
	model.UpdateTime = time.Now().UTC()

	err = PGConsoleUpdateNote(model)
	if err != nil {
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "更新笔记出错"))
		return
	}

	result := nemodels.NECodeOk.WithData(model.Uid)

	gctx.JSON(http.StatusOK, result)
}

func PGConsoleUpdateNote(model *MTNoteModel) error {
	sqlText := `update articles set title = :title,  body = :body, description = :description, 
	update_time = now() where uid = :uid;`

	sqlParams := map[string]interface{}{
		"uid":         model.Uid,
		"title":       model.Title,
		"body":        model.Body,
		"description": model.Description,
	}

	if _, err := datastore.NamedExec(sqlText, sqlParams); err != nil {
		return fmt.Errorf("PGConsoleUpdateNote: %w", err)
	}
	return nil
}
