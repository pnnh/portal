package notes

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"neutron/helpers"
	"neutron/services/datastore"
	"portal/business"
	"portal/models"
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
		gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "查询账号出错b"))
		return
	}
	if accountModel == nil {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("账号不存在"))
		return
	}
	selectResult, err := ConsoleSelectNotes(accountModel.Uid, channel, keyword, pageInt, sizeInt, lang)
	if err != nil {
		gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "查询笔记出错"))
		return
	}

	responseResult := models.CodeOk.WithData(selectResult)

	gctx.JSON(http.StatusOK, responseResult)
}

type MTConsoleNoteView struct {
	MTNoteModel
	Cid         string `json:"cid" db:"cid"`
	Lang        string `json:"lang" db:"lang"`
	ChannelName string `json:"channel_name" db:"channel_name"`
}

func ConsoleSelectNotes(owner, channel, keyword string, page int, size int, lang string) (*models.SelectResult[MTConsoleNoteView], error) {
	pagination := helpers.CalcPaginationByPage(page, size)
	baseSqlText := ` select a.*, c.name channel_name from articles as a join channels as c on a.channel = c.uid `
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
	var sqlResults []MTConsoleNoteView

	rows, err := datastore.NamedQuery(pageSqlText, pageSqlParams)
	if err != nil {
		return nil, fmt.Errorf("NamedQuery: %w", err)
	}
	if err = sqlx.StructScan(rows, &sqlResults); err != nil {
		return nil, fmt.Errorf("StructScan: %w", err)
	}

	resultRange := make([]MTConsoleNoteView, 0)
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

	selectData := &models.SelectResult[MTConsoleNoteView]{
		Page:  pagination.Page,
		Size:  pagination.Size,
		Count: countSqlResults[0].Count,
		Range: resultRange,
	}

	return selectData, nil
}

func ConsoleNoteGetHandler(gctx *gin.Context) {
	uid := gctx.Param("uid")
	lang := gctx.Query("lang")
	if uid == "" {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("uid不能为空"))
		return
	}
	accountModel, err := business.FindAccountFromCookie(gctx)
	if err != nil {
		logrus.Warnln("ConsoleNotesSelectHandler", err)
		gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "查询账号出错b"))
		return
	}
	if accountModel == nil {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("账号不存在"))
		return
	}
	selectResult, err := PGConsoleGetNote(accountModel.Uid, uid, lang)
	if err != nil {
		gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "查询笔记出错"))
		return
	}
	responseResult := models.CodeOk.WithData(selectResult)

	gctx.JSON(http.StatusOK, responseResult)
}

func ConsoleNoteDeleteHandler(gctx *gin.Context) {
	uid := gctx.Param("uid")
	lang := gctx.Query("lang")
	if uid == "" {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("uid不能为空"))
		return
	}
	accountModel, err := business.FindAccountFromCookie(gctx)
	if err != nil {
		logrus.Warnln("ConsoleNotesSelectHandler", err)
		gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "查询账号出错b"))
		return
	}
	if accountModel == nil {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("账号不存在"))
		return
	}
	err = PGConsoleDeleteNote(accountModel.Uid, uid, lang)
	if err != nil {
		gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "查询笔记出错"))
		return
	}
	responseResult := models.CodeOk.WithData(uid)

	gctx.JSON(http.StatusOK, responseResult)
}

func PGConsoleGetNote(owner, uid string, lang string) (*MTNoteModel, error) {
	if uid == "" {
		return nil, fmt.Errorf("PGConsoleGetNote uid is empty")
	}
	pageSqlText := ` select * from articles where (owner = :owner and (uid = :uid or (cid = :uid and lang = :lang))); `

	pageSqlParams := map[string]interface{}{
		"uid":   uid,
		"lang":  lang,
		"owner": owner,
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
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("uid不能为空"))
		return
	}
	accountModel, err := business.FindAccountFromCookie(gctx)
	if err != nil {
		logrus.Warnln("NoteUpdateHandler", err)
		gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "查询账号出错c"))
		return
	}
	if accountModel == nil || accountModel.IsAnonymous() {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("账号不存在或匿名用户不能修改笔记"))
		return
	}

	model := &MTNoteModel{}
	if err := gctx.ShouldBindJSON(model); err != nil {
		gctx.JSON(http.StatusOK, models.CodeError.WithError(err))
		return
	}
	if model.Title == "" || model.Body == "" {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("标题或内容不能为空"))
		return
	}
	oldModel, err := PGConsoleGetNote(accountModel.Uid, uid, "")
	if err != nil {
		gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "查询笔记出错"))
		return
	}
	if oldModel == nil {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("笔记不存在"))
		return
	}
	if oldModel.Owner != accountModel.Uid {
		gctx.JSON(http.StatusOK, models.CodeUnauthorized.WithMessage("没有权限修改该笔记"))
		return
	}

	model.Uid = uid
	model.UpdateTime = time.Now().UTC()

	err = PGConsoleUpdateNote(model)
	if err != nil {
		gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "更新笔记出错"))
		return
	}

	result := models.CodeOk.WithData(model.Uid)

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
