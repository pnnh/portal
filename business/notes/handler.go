package notes

import (
	"net/http"
	"strconv"
	"time"

	"neutron/helpers/jsonmap"
	nemodels "neutron/models"
	"neutron/services/datastore"
	"neutron/services/datetime"

	"neutron/helpers"
	"portal/business"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func NoteSelectHandler(gctx *gin.Context) {
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
	if lang == "" {
		lang = nemodels.DefaultLanguage
	}
	pagination, selectResult, err := SelectNotes(channel, keyword, pageInt, sizeInt, lang)
	if err != nil {
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "查询笔记出错"))
		return
	}

	resp := map[string]any{
		"page":  pagination.Page,
		"size":  pagination.Size,
		"count": pagination.Count,
		"range": selectResult,
	}

	//selectResponse := nemodels.NESelectResultToResponse(selectResult)
	responseResult := nemodels.NECodeOk.WithData(resp)

	gctx.JSON(http.StatusOK, responseResult)
}

func NoteConsoleInsertHandler(gctx *gin.Context) {
	accountModel, err := business.FindAccountFromCookie(gctx)
	if err != nil {
		logrus.Warnln("NoteConsoleInsertHandler", err)
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
	if inTitle == "" || inBody == "" || inLang == "" || jsonMap.Err != nil {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("标题或内容不能为空3"))
		return
	}

	if !nemodels.IsValidLanguage(inLang) {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("Lang参数错误"))
		return
	}
	modelUid := helpers.MustUuid()
	nowTime := time.Now()
	dataRow := datastore.NewDataRow()
	dataRow.SetString("uid", modelUid)

	dataRow = dataRow.SetStringChainFrom("title", jsonMap).
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

	err = PGConsoleInsertNote(dataRow)
	if err != nil {
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "插入笔记出错"))
		return
	}

	result := nemodels.NECodeOk.WithData(map[string]any{
		"changes": 1,
		"uid":     modelUid,
	})

	gctx.JSON(http.StatusOK, result)
}

func NoteGetHandler(gctx *gin.Context) {
	uid := gctx.Param("uid")
	lang := gctx.Query("lang")
	if uid == "" {
		gctx.JSON(http.StatusOK, nemodels.NECodeError.WithMessage("uid不能为空"))
		return
	}
	noteTable, err := PGGetNote(uid, lang)
	if err != nil || noteTable == nil {
		gctx.JSON(http.StatusOK, nemodels.NEErrorResultMessage(err, "查询笔记出错"))
		return
	}
	noteModel := noteTable.ToModel()
	viewModel := noteModel.ToViewModel()
	responseResult := nemodels.NECodeOk.WithLocalData(lang, viewModel)

	gctx.JSON(http.StatusOK, responseResult)
}
