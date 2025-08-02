package notes

import (
	"database/sql"
	"fmt"
	"github.com/iancoleman/strcase"
	nemodels "neutron/models"
	"neutron/services/maputil"
	"neutron/services/strutil"
	"reflect"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"neutron/helpers"
	"neutron/services/datastore"
)

type MTNoteMatter struct {
	Cls         string `json:"cls"`
	Uid         string `json:"uid"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type MTNoteTable struct {
	Uid         string         `json:"uid"`
	Title       string         `json:"title"`
	Header      string         `json:"header"`
	Body        string         `json:"body"`
	Description string         `json:"description"`
	Keywords    string         `json:"keywords"`
	Status      int            `json:"status"`
	Cover       sql.NullString `json:"-"`
	Owner       string         `json:"-"`
	Channel     string         `json:"channel"`
	Discover    int            `json:"discover"`
	Partition   sql.NullString `json:"-"`
	CreateTime  time.Time      `json:"create_time" db:"create_time"`
	UpdateTime  time.Time      `json:"update_time" db:"update_time"`
	Version     sql.NullString `json:"-"`
	Build       sql.NullString `json:"-"`
	Url         sql.NullString `json:"-"`
	Branch      sql.NullString `json:"-"`
	Commit      sql.NullString `json:"-" db:"commit"`
	CommitTime  sql.NullTime   `json:"-" db:"commit_time"`
	RepoPath    sql.NullString `json:"-" db:"repo_path"`
	RepoId      sql.NullString `json:"-" db:"repo_id"`
	Cid         string         `json:"cid" db:"cid"`
	Lang        string         `json:"lang" db:"lang"`
	Nid         int64          `json:"nid" db:"nid" insert:"skip"`
	Dc          string         `json:"-" db:"dc"`
	Name        string         `json:"name" db:"name"`
}

func (t *MTNoteTable) ToModel() *MTNoteModel {
	return &MTNoteModel{
		MTNoteTable: *t,
		Partition:   t.Partition.String,
		Version:     t.Version.String,
		Build:       t.Build.String,
		Url:         t.Url.String,
		Branch:      t.Branch.String,
		Commit:      t.Commit.String,
		CommitTime:  t.CommitTime.Time,
		RepoPath:    t.RepoPath.String,
		RepoId:      t.RepoId.String,
	}
}

type MTNoteModel struct {
	MTNoteTable
	Partition  string    `json:"-"`
	Version    string    `json:"-"`
	Build      string    `json:"-"`
	Url        string    `json:"-"`
	Branch     string    `json:"-"`
	Commit     string    `json:"-" db:"commit"`
	CommitTime time.Time `json:"-" db:"commit_time"`
	RepoPath   string    `json:"-" db:"repo_path"`
	RepoId     string    `json:"-" db:"repo_id"`
}

func (m MTNoteModel) ToViewModel() interface{} {
	view := &MTNoteView{
		MTNoteModel: m,
	}
	//if m.Cid.Valid {
	//	view.Cid = m.Cid.String
	//}
	//if m.Lang.Valid {
	//	view.Lang = m.Lang.String
	//}
	view.Cid = m.Cid
	view.Lang = m.Lang
	return view

}

type MTNoteView struct {
	MTNoteModel
	Cid  string `json:"cid" db:"cid"`
	Lang string `json:"lang" db:"lang"`
}

func ReflectColumns(s interface{}) (map[string]any, error) {
	v := reflect.ValueOf(s)
	t := reflect.TypeOf(s)

	// 如果是指针，取其元素
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("getStructFields kind is not struct")
	}
	columnMap := make(map[string]any)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		fmt.Printf("成员名: %s, 类型: %s, 类型名称: %s, 值: %v\n",
			field.Name,
			field.Type.Kind(),      // 基本类型，如 string、int、struct 等
			field.Type.Name(),      // 类型名称，如 int、string、自定义类型名
			fieldValue.Interface(), // 字段值
		)
		colName := strcase.ToSnake(field.Name)
		dbTag := field.Tag.Get("db")
		if dbTag == "-" {
			continue
		} else if dbTag != "" {
			colName = dbTag
		}
		insertTag := field.Tag.Get("insert")
		if insertTag == "skip" {
			continue
		}
		var colValue any
		switch val := fieldValue.Interface().(type) {
		case sql.NullString:
			if val.Valid {
				colValue = val.String
			}
		default:
			colValue = val
		}
		columnMap[colName] = colValue
	}
	return columnMap, nil
}

func PGConsoleInsertNote(model *MTNoteModel) error {
	columnMap, err := ReflectColumns(&model.MTNoteTable)
	if err != nil {
		return fmt.Errorf("PGConsoleInsertNote ReflectColumns: %w", err)
	}

	colNames := maputil.StringMapKeys(columnMap)
	colText := strings.Join(colNames, ", ")
	colPlaceholders := strutil.JoinStringsFunc(colNames, func(s string) string {
		return fmt.Sprintf(":%s, ", s)
	}, func(s string) string {
		return strings.TrimRight(s, ", ")
	})

	sqlText := fmt.Sprintf(`insert into articles(%s)
values(%s)
on conflict (uid)
do nothing;`, colText, colPlaceholders)

	_, err = datastore.NamedExec(sqlText, columnMap)
	if err != nil {
		return fmt.Errorf("PGConsoleInsertNote: %w", err)
	}
	return nil
}

func SelectNotes(channel, keyword string, page int, size int, lang string) (*nemodels.NESelectResult[MTNoteModel], error) {
	pagination := helpers.CalcPaginationByPage(page, size)
	baseSqlText := ` select * from articles `
	baseSqlParams := map[string]interface{}{}

	whereText := ` where status = 1 `
	if keyword != "" {
		whereText += ` and (title like :keyword or description like :keyword) `
		baseSqlParams["keyword"] = "%" + keyword + "%"
	}
	if channel != "" {
		whereText += ` and channel = :channel `
		baseSqlParams["channel"] = channel
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
	var sqlResults []MTNoteModel

	rows, err := datastore.NamedQuery(pageSqlText, pageSqlParams)
	if err != nil {
		return nil, fmt.Errorf("NamedQuery: %w", err)
	}
	if err = sqlx.StructScan(rows, &sqlResults); err != nil {
		return nil, fmt.Errorf("StructScan: %w", err)
	}

	resultRange := make([]MTNoteModel, 0)
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

	selectData := &nemodels.NESelectResult[MTNoteModel]{
		Page:  pagination.Page,
		Size:  pagination.Size,
		Count: countSqlResults[0].Count,
		Range: resultRange,
	}

	return selectData, nil
}

// PGGetNote 获取单个笔记信息
// obsolete 处理向后兼容逻辑。早期是通过单个uid查询，后期可以通过cid和lang查询。
func PGGetNote(uid string, lang string) (*MTNoteModel, error) {
	if uid == "" {
		return nil, fmt.Errorf("PGGetNote uid is empty")
	}
	pageSqlText := ` select * from articles where status = 1 and (uid = :uid or (cid = :uid and lang = :lang)); `

	pageSqlParams := map[string]interface{}{
		"uid":  uid,
		"lang": lang,
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

type MTNoteFileModel struct {
	Title       string `json:"title"`
	Path        string `json:"path"`
	IsDir       bool   `json:"is_dir"`
	IsText      bool   `json:"is_text"`
	IsImage     bool   `json:"is_image"`
	StoragePath string `json:"storage_path"`
}
