package models

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"portal/neutron/services/datastore"
)

type AccessTokenModel struct {
	Pk         string    `json:"urn"`
	CreateTime time.Time `json:"create_time" db:"create_time"`
	UpdateTime time.Time `json:"update_time" db:"update_time"`
	Signature  string    `json:"signature"`
	Content    string    `json:"content"`
}

func PutAccessToken(model *AccessTokenModel) error {
	sqlText := `insert into portal.access_token(pk, create_time, update_time, signature, content)
		values(:urn, :create_time, :update_time, :signature, :content);`

	sqlParams := map[string]interface{}{
		"urn":         model.Pk,
		"create_time": model.CreateTime,
		"update_time": model.UpdateTime,
		"signature":   model.Signature,
		"content":     model.Content,
	}

	_, err := datastore.NamedExec(sqlText, sqlParams)
	if err != nil {
		return fmt.Errorf("PutSession: %w", err)
	}
	return nil
}

func DeleteAccessToken(pk string) error {
	sqlText := `delete from portal.access_token where pk = :urn or signature = :urn;`

	sqlParams := map[string]interface{}{
		"urn": pk,
	}

	_, err := datastore.NamedExec(sqlText, sqlParams)
	if err != nil {
		return fmt.Errorf("DeleteAccessToken: %w", err)
	}
	return nil
}

func GetAccessToken(pk string) (*AccessTokenModel, error) {
	sqlText := `select * from portal.access_token where pk = :urn or signature = :urn;`

	sqlParams := map[string]interface{}{"urn": pk}
	var sqlResults []*AccessTokenModel

	rows, err := datastore.NamedQuery(sqlText, sqlParams)
	if err != nil {
		return nil, fmt.Errorf("NamedQuery: %w", err)
	}
	if err = sqlx.StructScan(rows, &sqlResults); err != nil {
		return nil, fmt.Errorf("StructScan: %w", err)
	}

	for _, v := range sqlResults {
		return v, nil
	}

	return nil, nil
}
