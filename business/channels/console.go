package channels

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

func ConsoleChannelGetHandler(gctx *gin.Context) {
	uid := gctx.Param("uid")
	lang := gctx.Query("lang")
	if uid == "" {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("uid不能为空"))
		return
	}
	accountModel, err := business.FindAccountFromCookie(gctx)
	if err != nil {
		logrus.Warnln("ConsoleChannelsSelectHandler", err)
		gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "查询账号出错b"))
		return
	}
	if accountModel == nil {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("账号不存在"))
		return
	}
	selectResult, err := PGConsoleGetChannel(accountModel.Uid, uid, lang)
	if err != nil {
		gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "查询频道出错"))
		return
	}
	var modelData any
	if selectResult != nil {
		modelData = selectResult.ToViewModel()
	}
	responseResult := models.CodeOk.WithData(modelData)

	gctx.JSON(http.StatusOK, responseResult)
}

func PGConsoleGetChannel(owner, uid, lang string) (*MTChannelModel, error) {
	pageSqlText := ` select * from channels where (owner = :owner and (uid = :uid) or (cid = :uid and lang = :lang)); `
	pageSqlParams := map[string]interface{}{
		"uid":   uid,
		"lang":  lang,
		"owner": owner,
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

func ConsoleChannelInsertHandler(gctx *gin.Context) {
	accountModel, err := business.FindAccountFromCookie(gctx)
	if err != nil {
		logrus.Warnln("ChannelConsoleInsertHandler", err)
		gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "查询账号出错c"))
		return
	}
	if accountModel == nil || accountModel.IsAnonymous() {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("账号不存在或匿名用户不能发布频道"))
		return
	}

	model := &MTChannelView{}
	if err := gctx.ShouldBindJSON(model); err != nil {
		gctx.JSON(http.StatusOK, models.CodeError.WithError(err))
		return
	}
	if model.Name == "" {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("标题或内容不能为空"))
		return
	}
	if model.Lang == "" || (model.Lang != business.LangZh && model.Lang != business.LangEn) {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("Lang参数错误"))
		return
	}

	model.Uid = helpers.MustUuid()
	model.Owner = accountModel.Uid
	model.CreateTime = time.Now().UTC()
	model.UpdateTime = time.Now().UTC()
	model.Status = 0 // 待审核
	model.Cid = model.Uid

	err = PGConsoleInsertChannel(model)
	if err != nil {
		gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "插入频道出错"))
		return
	}

	result := models.CodeOk.WithData(model.Uid)

	gctx.JSON(http.StatusOK, result)
}

func PGConsoleInsertChannel(model *MTChannelView) error {
	sqlText := `insert into channels(uid, name, title, description, image, status, create_time, update_time, lang, owner)
values(:uid, :name, :title, :description, :image, 0, now(), now(), :lang, :owner)
on conflict (uid)
do nothing;`

	sqlParams := map[string]interface{}{
		"uid":         model.Uid,
		"name":        model.Name,
		"title":       model.Title,
		"description": model.Description,
		"lang":        model.Lang,
		"image":       model.Image,
		"owner":       model.Owner,
	}

	_, err := datastore.NamedExec(sqlText, sqlParams)
	if err != nil {
		return fmt.Errorf("PGConsoleInsertChannel: %w", err)
	}
	return nil
}

func ConsoleChannelUpdateHandler(gctx *gin.Context) {
	uid := gctx.Param("uid")
	if uid == "" {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("uid不能为空"))
		return
	}
	accountModel, err := business.FindAccountFromCookie(gctx)
	if err != nil {
		logrus.Warnln("ChannelUpdateHandler", err)
		gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "查询账号出错c"))
		return
	}
	if accountModel == nil || accountModel.IsAnonymous() {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("账号不存在或匿名用户不能修改频道"))
		return
	}

	model := &MTChannelView{}
	if err := gctx.ShouldBindJSON(model); err != nil {
		gctx.JSON(http.StatusOK, models.CodeError.WithError(err))
		return
	}
	if model.Title == "" {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("标题不能为空"))
		return
	}
	if model.Name == "" {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("名称不能为空"))
		return
	}
	oldModel, err := PGConsoleGetChannel(accountModel.Uid, uid, "")
	if err != nil {
		gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "查询频道出错"))
		return
	}
	if oldModel == nil {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("频道不存在"))
		return
	}
	if oldModel.Owner != accountModel.Uid {
		gctx.JSON(http.StatusOK, models.CodeUnauthorized.WithMessage("没有权限修改该频道"))
		return
	}

	model.Uid = uid
	model.UpdateTime = time.Now().UTC()

	err = PGConsoleUpdateChannel(model)
	if err != nil {
		gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "更新频道出错"))
		return
	}

	result := models.CodeOk.WithData(model.Uid)

	gctx.JSON(http.StatusOK, result)
}

func PGConsoleUpdateChannel(model *MTChannelView) error {
	sqlText := `update channels set name = :name, title = :title, description = :description, lang = :lang, image = :image,
	update_time = now() where uid = :uid;`

	sqlParams := map[string]interface{}{
		"uid":         model.Uid,
		"name":        model.Name,
		"title":       model.Title,
		"description": model.Description,
		"lang":        model.Lang,
		"image":       model.Image,
	}

	if _, err := datastore.NamedExec(sqlText, sqlParams); err != nil {
		return fmt.Errorf("PGConsoleUpdateChannel: %w", err)
	}
	return nil
}

func ConsoleChannelSelectHandler(gctx *gin.Context) {
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
		sizeInt = 100
	}

	accountModel, err := business.FindAccountFromCookie(gctx)
	if err != nil {
		logrus.Warnln("ConsoleChannelsSelectHandler", err)
		gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "查询账号出错b"))
		return
	}
	if accountModel == nil {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("账号不存在"))
		return
	}
	selectResult, err := ConsoleSelectChannels(accountModel.Uid, keyword, pageInt, sizeInt, lang)
	if err != nil {
		gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "查询频道出错"))
		return
	}

	selectResponse := models.SelectResultToResponse(selectResult)
	responseResult := models.CodeOk.WithData(selectResponse)

	gctx.JSON(http.StatusOK, responseResult)
}

func ConsoleSelectChannels(owner, keyword string, page int, size int, lang string) (*models.SelectResult[MTChannelModel], error) {
	pagination := helpers.CalcPaginationByPage(page, size)
	baseSqlText := ` select * from channels `
	baseSqlParams := map[string]interface{}{}

	whereText := ` where owner = :owner `
	baseSqlParams["owner"] = owner
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
		return nil, fmt.Errorf("查询频道总数有误，数据为空")
	}

	selectData := &models.SelectResult[MTChannelModel]{
		Page:  pagination.Page,
		Size:  pagination.Size,
		Count: countSqlResults[0].Count,
		Range: resultRange,
	}

	return selectData, nil
}

func ConsoleChannelDeleteHandler(gctx *gin.Context) {
	uid := gctx.Param("uid")
	lang := gctx.Query("lang")
	if uid == "" {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("uid不能为空"))
		return
	}
	accountModel, err := business.FindAccountFromCookie(gctx)
	if err != nil {
		logrus.Warnln("ConsoleChannelsSelectHandler", err)
		gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "查询账号出错b"))
		return
	}
	if accountModel == nil {
		gctx.JSON(http.StatusOK, models.CodeError.WithMessage("账号不存在"))
		return
	}
	err = PGConsoleDeleteChannel(accountModel.Uid, uid, lang)
	if err != nil {
		gctx.JSON(http.StatusOK, models.ErrorResultMessage(err, "查询频道出错"))
		return
	}
	responseResult := models.CodeOk.WithData(uid)

	gctx.JSON(http.StatusOK, responseResult)
}

func PGConsoleDeleteChannel(owner, uid string, lang string) error {
	if uid == "" {
		return fmt.Errorf("PGConsoleGetChannel uid is empty")
	}
	pageSqlText := ` delete from channels where (owner = :owner and (uid = :uid or (cid = :uid and lang = :lang))); `

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
