package notes

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"portal/neutron/services/datastore"
)

type MTViewerModel struct {
	Uid        string         `json:"uid"`
	Title      string         `json:"title"`
	Source     sql.NullString `json:"source"`
	Target     string         `json:"target"`
	CreateTime time.Time      `json:"create_time" db:"create_time"`
	UpdateTime time.Time      `json:"update_time" db:"update_time"`
	Address    string         `json:"address"`
	Class      string         `json:"class"`
}

var ErrViewerLogExists = fmt.Errorf("viewer log exists")

func PGInsertViewer(viewerModels ...*MTViewerModel) (opErr error, itemErrs map[string]error) {
	sqlTx, err := datastore.NewTranscation()
	if err != nil {
		return fmt.Errorf("PGViewerNote: %w", err), nil
	}
	defer func() {
		if p := recover(); p != nil {
			err = sqlTx.Rollback()
			if err != nil {
				err = fmt.Errorf("PGViewerNote Rollback: %w\n%v", err, p)
			}
		}
	}()

	nowYear, nowMonth, nowDay := time.Now().AddDate(0, 0, -1).Date()
	nowDate := time.Date(nowYear, nowMonth, nowDay, 0, 0, 0, 0, time.UTC)
	for _, model := range viewerModels {

		queryText := `
	select * from viewer where class = :class and target = :target and update_time > :nowDate 
	                       and (source = :source or address = :address) limit 1;
`
		queryParams := map[string]interface{}{
			"source":  model.Source,
			"target":  model.Target,
			"address": model.Address,
			"class":   model.Class,
			"nowDate": nowDate,
		}
		queryRows, err := sqlTx.NamedQuery(queryText, queryParams)
		if err != nil {
			itemErrs[model.Uid] = fmt.Errorf("PGViewerNote query: %w", err)
			continue
		}
		var queryResults []*MTViewerModel
		if err = sqlx.StructScan(queryRows, &queryResults); err != nil {
			itemErrs[model.Uid] = fmt.Errorf("StructScan: %w", err)
			continue
		}
		if len(queryResults) > 0 {
			viewerLog := queryResults[0]
			if viewerLog.UpdateTime.After(time.Now().Add(-1 * time.Hour * 24)) {
				itemErrs[model.Uid] = ErrViewerLogExists
				continue
			}
		}
		if err = queryRows.Close(); err != nil {
			itemErrs[model.Uid] = fmt.Errorf("PGViewerNote query close: %w", err)
			continue
		}

		viewerText := `
insert into viewer(uid, source, target, create_time, update_time, title, address, class)
values(:uid, :source, :target, :create_time, :update_time, :title, :address, :class)
on conflict (uid)
do update set title=excluded.title, source=excluded.source, target=excluded.target, address=excluded.address, 
    update_time = now();
`
		sqlParams := map[string]interface{}{
			"uid":         model.Uid,
			"source":      model.Source,
			"target":      model.Target,
			"create_time": model.CreateTime,
			"update_time": model.UpdateTime,
			"title":       model.Title,
			"address":     model.Address,
			"class":       model.Class,
		}
		viewerRows, err := sqlTx.NamedQuery(viewerText, sqlParams)
		if err != nil {
			itemErrs[model.Uid] = fmt.Errorf("PGViewerNote viewer: %w", err)
			continue
		}
		if err = viewerRows.Close(); err != nil {
			itemErrs[model.Uid] = fmt.Errorf("PGViewerNote viewer close: %w", err)
			continue
		}

		err = updateObjectDiscover(sqlTx, model)
		if err != nil {
			itemErrs[model.Uid] = fmt.Errorf("PGViewerNote discover: %w", err)
			continue
		}
	}

	if err = sqlTx.Commit(); err != nil {
		err = fmt.Errorf("PGViewerNote Commit: %w", err)
	}

	return err, itemErrs
}

func updateObjectDiscover(sqlTx *datastore.SqlxTransaction, model *MTViewerModel) error {

	discoverSqlText := `update articles set discover = COALESCE(discover, 0) + 1 where uid = :uid;`
	if model.Class == "comment" {
		discoverSqlText = `update comments set discover = COALESCE(discover, 0) + 1 where uid = :uid;`
	}
	discoverSqlParams := map[string]interface{}{
		"uid": model.Target,
	}
	discoverRows, err := sqlTx.NamedQuery(discoverSqlText, discoverSqlParams)
	if err != nil {
		return fmt.Errorf("PGViewerNote discover: %w", err)
	}
	if err = discoverRows.Close(); err != nil {
		return fmt.Errorf("PGViewerNote discover close: %w", err)
	}
	return nil
}
