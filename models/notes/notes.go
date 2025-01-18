package notes

import (
	"fmt"
	"portal/neutron/services/datastore"
	"time"
)

type MTNoteMatter struct {
	Cls         string `json:"cls"`
	Uid         string `json:"uid"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type MTNoteModel struct {
	Uid         string    `json:"uid"`
	Title       string    `json:"title"`
	Body        string    `json:"body"`
	Description string    `json:"description"`
	CreateTime  time.Time `json:"create_time" db:"create_time"`
	UpdateTime  time.Time `json:"update_time" db:"update_time"`
}

func PGInsertNote(model *MTNoteModel) error {
	sqlText := `insert into articles(uid, title, header, body, description, create_time, update_time)
values(:uid, :title, :header, :body, :description, now(), now())
on conflict (uid)
do update set title=excluded.title, body=excluded.body, update_time = now();`

	sqlParams := map[string]interface{}{
		"uid":         model.Uid,
		"title":       model.Title,
		"header":      "MTNote",
		"body":        model.Body,
		"description": model.Description,
	}

	_, err := datastore.NamedExec(sqlText, sqlParams)
	if err != nil {
		return fmt.Errorf("PGInsertNote: %w", err)
	}
	return nil
}
