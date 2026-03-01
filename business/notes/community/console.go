package community

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/pnnh/neutron/helpers/jsonmap"
	nemodels "github.com/pnnh/neutron/models"
	"github.com/pnnh/neutron/services/datetime"
	"portal/business/notes"

	"github.com/pnnh/neutron/helpers"
	"github.com/pnnh/neutron/services/datastore"
	"portal/business"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type ConsoleNotesHandler struct {
}

func (h *ConsoleNotesHandler) RegisterRouter(router *gin.Engine) {
	//router.GET("/portal/console/community/articles", h.HandleSelect)
	//router.GET("/portal/:lang/console/articles", h.HandleSelect)
	//router.POST("/portal/console/community/articles", h.HandleInsert)
	//router.GET("/portal/console/community/articles/:uid", h.HandleGet)
	//router.GET("/portal/:lang/console/articles/:uid", h.HandleGet)
	//router.POST("/portal/console/community/articles/:uid", h.HandleUpdate)
	//router.DELETE("/portal/console/community/articles/:uid", h.HandleDelete)
}

func (h *ConsoleNotesHandler) HandleSelect(gctx *gin.Context) {
	keyword := gctx.Query("keyword")
	page := gctx.Query("page")
	size := gctx.Query("size")
	channel := gctx.Query("channel")
	lang := gctx.Query("lang")
	pageInt, err := strconv.Atoi(page)
	action := gctx.Query("action")
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

	if action == "get" {
		noteRow, err := h.PGConsoleGetNote(accountModel.Uid, keyword)
		if err != nil {
			gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "查询笔记出错"))
			return
		}

		outView, err := h.consoleNoteGetOutView(noteRow)
		if err != nil {
			gctx.JSON(http.StatusOK, nemodels.NECodeError.WithError(err))
			return
		}

		resp := map[string]any{
			"page":  1,
			"size":  1,
			"count": 1,
			"range": []any{outView},
		}

		responseResult := nemodels.NECodeOk.WithData(resp)

		gctx.JSON(http.StatusOK, responseResult)
		return
	}

	pagination, selectResult, err := h.PGSelectNotes(accountModel.Uid, channel, keyword, pageInt, sizeInt, lang)
	if err != nil {
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "查询笔记出错"))
		return
	}
	respView := make([]map[string]interface{}, 0)
	for _, v := range selectResult {
		outView, err := h.consoleNoteGetOutView(v)
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

func (h *ConsoleNotesHandler) PGSelectNotes(owner, channel, keyword string, page int, size int, lang string) (*helpers.Pagination,
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
			logrus.Warnf("rows.Close: %v", closeErr)
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

func (h *ConsoleNotesHandler) PGConsoleInsertNote(dataRow *datastore.DataRow) error {
	sqlText := `insert into articles(uid, title, header, body, create_time, update_time, keywords, description, status, 
	cover, owner, channel, discover, partition, version, build, url, branch, commit, commit_time, relative_path, repo_id, 
	lang, name, checksum, syncno, repo_first_commit)
values(:uid, :title, :header, :body, :create_time, :update_time, :keywords, :description, :status, :cover, :owner, 
	:channel, :discover, :partition, :version, :build, :url, :branch, :commit, :commit_time, :relative_path, :repo_id, 
	:lang, :name, :checksum, :syncno, :repo_first_commit)
on conflict (uid)
do update set title = excluded.title,
              header = excluded.header,
              body = excluded.body,
              update_time = excluded.update_time,
              keywords = excluded.keywords,
              description = excluded.description,
              status = excluded.status,
              cover = excluded.cover,
              owner = excluded.owner,
              channel = excluded.channel,
              partition = excluded.partition,
              version = excluded.version,
              build = excluded.build,
              url = excluded.url,
              branch = excluded.branch,
              commit = excluded.commit,
              commit_time = excluded.commit_time,
              relative_path = excluded.relative_path,
              repo_id = excluded.repo_id,
              lang = excluded.lang,
              name = excluded.name,
              checksum = excluded.checksum,
              syncno = excluded.syncno,
			  repo_first_commit=excluded.repo_first_commit
where articles.owner = :owner;`

	paramsMap := dataRow.InnerMap()

	_, err := datastore.NamedExec(sqlText, paramsMap)
	if err != nil {
		return fmt.Errorf("PGConsoleInsertNote: %w", err)
	}
	return nil
}

func (h *ConsoleNotesHandler) consoleNoteGetOutView(dataRow *datastore.DataRow) (map[string]interface{}, error) {
	outView := make(map[string]interface{})
	outView["uid"] = dataRow.GetString("uid")
	outView["title"] = dataRow.GetString("title")
	outView["header"] = dataRow.GetString("header")
	outView["body"] = dataRow.GetString("body")
	outView["description"] = dataRow.GetStringOrEmpty("description")
	outView["keywords"] = dataRow.GetStringOrDefault("keywords", "")
	outView["status"] = dataRow.GetInt("status")
	outView["cover"] = dataRow.GetStringOrDefault("cover", "")
	outView["owner"] = dataRow.GetNullString("owner")
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

func (h *ConsoleNotesHandler) PGConsoleGetNote(owner, uid string) (*datastore.DataRow, error) {
	if uid == "" {
		return nil, fmt.Errorf("PGConsoleGetNote uid is empty")
	}

	pageSqlParams := map[string]interface{}{
		"owner": owner,
	}
	var pageSqlText string

	pageSqlText = ` select * from articles where owner = :owner and uid = :uid; `
	pageSqlParams["uid"] = uid

	rows, err := datastore.NamedQuery(pageSqlText, pageSqlParams)
	if err != nil {
		return nil, fmt.Errorf("NamedQuery: %w", err)
	}

	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			logrus.Warnf("rows.Close: %v", closeErr)
		}
	}()

	for rows.Next() {
		rowMap := make(map[string]interface{})
		if err := rows.MapScan(rowMap); err != nil {
			return nil, fmt.Errorf("MapScan: %w", err)
		}
		tableMap := datastore.MapToDataRow(rowMap)
		if tableMap.Err != nil {
			return nil, fmt.Errorf("MapScan2: %w", tableMap.Err)
		}
		return tableMap, nil
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return nil, nil
}

func (h *ConsoleNotesHandler) PGConsoleDeleteNote(owner, uid string, lang string) error {
	if uid == "" {
		return fmt.Errorf("PGConsoleGetNote uid is empty")
	}
	pageSqlText := ` delete from public.articles where owner = :owner and uid = :uid; `

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

func (h *ConsoleNotesHandler) HandleUpdate(gctx *gin.Context) {
	uid := gctx.Param("uid")
	if uid == "" {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("uid不能为空"))
		return
	}
	action := gctx.Query("action")

	accountModel, err := business.FindAccountFromCookie(gctx)
	if err != nil {
		logrus.Warnln("HandleInsert", err)
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "查询账号出错c"))
		return
	}
	if accountModel == nil || accountModel.IsAnonymous() {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("账号不存在或匿名用户不能发布笔记"))
		return
	}

	jsonMap := jsonmap.NewJsonMap()

	if err := gctx.ShouldBind(jsonMap.InnerMapPtr()); err != nil {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithError(err))
		return
	}

	inTitle := jsonMap.WillGetString("title")
	inBody := jsonMap.WillGetString("body")
	inLang := jsonMap.WillGetString("lang")
	if action != "delete" {
		if inTitle == "" || inBody == "" || inLang == "" || jsonMap.Err != nil {
			gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("标题或内容不能为空3"))
			return
		}
		if !nemodels.IsValidLanguage(inLang) {
			gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("Lang参数错误"))
			return
		}
	}
	nowTime := time.Now()
	dataRow := datastore.NewDataRow()

	// 新增记录
	if uid == helpers.EmptyUuid() {
		uid = helpers.MustUuid()
		dataRow.SetString("uid", uid)
	} else if action == "delete" {
		err = h.PGConsoleDeleteNote(accountModel.Uid, uid, inLang)

		if err != nil {
			gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "查询笔记出错"))
			return
		}

		result := nemodels.NECodeOk.WithData(map[string]any{
			"changes": 1,
			"uid":     uid,
		})

		gctx.JSON(http.StatusOK, result)
		return
	}

	dataRow = dataRow.SetStringChain("uid", uid).SetStringChainFrom("title", jsonMap).
		SetStringChainFrom("header", jsonMap).SetStringChainFrom("body", jsonMap).
		SetNullStringChainFrom("description", jsonMap).SetNullStringChainFrom("keywords", jsonMap).
		SetIntChain("status", 0).SetStringChainFrom("cover", jsonMap).
		SetNullUuidStringChain("owner", accountModel.Uid).SetNullUuidStringChainFrom("channel", jsonMap).
		SetIntChain("discover", 0).SetNullUuidStringChainFrom("partition", jsonMap).
		SetNullTimeChain("create_time", nowTime).SetNullTimeChain("update_time", nowTime).
		SetNullStringChainFrom("version", jsonMap).SetNullStringChainFrom("build", jsonMap).
		SetNullStringChainFrom("url", jsonMap).SetNullStringChainFrom("branch", jsonMap).
		SetNullStringChainFrom("commit", jsonMap).SetNullTimeChain("commit_time", datetime.NullTime).
		SetNullStringChainFrom("relative_path", jsonMap).SetNullUuidStringChainFrom("repo_id", jsonMap).
		SetStringChainFrom("lang", jsonMap).SetNullStringChainFrom("name", jsonMap).
		SetNullStringChainFrom("checksum", jsonMap).SetNullStringChainFrom("syncno", jsonMap).
		SetNullStringChainFrom("repo_first_commit", jsonMap)

	if dataRow.Err != nil {
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(dataRow.Err, "参数错误2"))
		return
	}

	err = h.PGConsoleInsertNote(dataRow)
	if err != nil {
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "插入笔记出错"))
		return
	}

	result := nemodels.NECodeOk.WithData(map[string]any{
		"changes": 1,
		"uid":     uid,
	})

	gctx.JSON(http.StatusOK, result)
}

func (h *ConsoleNotesHandler) PGConsoleUpdateNote(model *notes.MTNoteModel) error {
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
