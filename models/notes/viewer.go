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
}

var ErrViewerLogExists = fmt.Errorf("viewer log exists")

func PGInsertViewer(model *MTViewerModel) (err error) {
	sqlTx, err := datastore.NewTranscation()
	if err != nil {
		return fmt.Errorf("PGViewerNote: %w", err)
	}
	defer func() {
		if p := recover(); p != nil || err != nil {
			err = sqlTx.Rollback()
			if err != nil {
				err = fmt.Errorf("PGViewerNote Rollback: %w\n%v", err, p)
			}
		}
	}()

	queryText := `
	select * from viewer where target = :target and (source = :source or address = :address) limit 1;
`
	queryParams := map[string]interface{}{
		"source":  model.Source,
		"target":  model.Target,
		"address": model.Address,
	}
	queryRows, err := sqlTx.NamedQuery(queryText, queryParams)
	if err != nil {
		err = fmt.Errorf("PGViewerNote query: %w", err)
		return err
	}
	var queryResults []*MTViewerModel
	if err = sqlx.StructScan(queryRows, &queryResults); err != nil {
		err = fmt.Errorf("StructScan: %w", err)
		return err
	}
	if len(queryResults) > 0 {
		viewerLog := queryResults[0]
		if viewerLog.UpdateTime.After(time.Now().Add(-1 * time.Hour * 24)) {
			return ErrViewerLogExists
		}
	}
	if err = queryRows.Close(); err != nil {
		err = fmt.Errorf("PGViewerNote query close: %w", err)
		return err
	}

	viewerText := `
insert into viewer(uid, source, target, create_time, update_time, title, address)
values(:uid, :source, :target, :create_time, :update_time, :title, :address)
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
	}
	viewerRows, err := sqlTx.NamedQuery(viewerText, sqlParams)
	if err != nil {
		err = fmt.Errorf("PGViewerNote viewer: %w", err)
		return err
	}
	if err = viewerRows.Close(); err != nil {
		err = fmt.Errorf("PGViewerNote viewer close: %w", err)
		return err
	}

	discoverSqlText := `update articles set discover = COALESCE(discover, 0) + 1 where uid = :uid;`
	discoverSqlParams := map[string]interface{}{
		"uid": model.Target,
	}
	discoverRows, err := sqlTx.NamedQuery(discoverSqlText, discoverSqlParams)
	if err != nil {
		err = fmt.Errorf("PGViewerNote discover: %w", err)
		return err
	}
	if err = discoverRows.Close(); err != nil {
		err = fmt.Errorf("PGViewerNote discover close: %w", err)
		return err
	}

	if err = sqlTx.Commit(); err != nil {
		err = fmt.Errorf("PGViewerNote Commit: %w", err)
	}

	return err
}
