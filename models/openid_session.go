package models

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"portal/neutron/services/datastore"
)

type OpenidSessionModel struct {
	Pk         string    `json:"urn"`
	CreateTime time.Time `json:"create_time" db:"create_time"`
	UpdateTime time.Time `json:"update_time" db:"update_time"`
	Code       string    `json:"code"`
	Content    string    `json:"content"`
}

func PutOpenidSession(model *OpenidSessionModel) error {
	sqlText := `insert into portal.openid_session(pk, create_time, update_time, code, content)
		values(:urn, :create_time, :update_time, :code, :content);`

	sqlParams := map[string]interface{}{
		"urn":         model.Pk,
		"create_time": model.CreateTime,
		"update_time": model.UpdateTime,
		"code":        model.Code,
		"content":     model.Content,
	}

	_, err := datastore.NamedExec(sqlText, sqlParams)
	if err != nil {
		return fmt.Errorf("PutSession: %w", err)
	}
	return nil
}

func DeleteOpenidSession(pk string) error {
	sqlText := `delete from portal.openid_session where pk = :urn or code = :urn;`

	sqlParams := map[string]interface{}{
		"urn": pk,
	}

	_, err := datastore.NamedExec(sqlText, sqlParams)
	if err != nil {
		return fmt.Errorf("DeleteOpenidSession: %w", err)
	}
	return nil
}

func GetOpenidSession(pk string) (*OpenidSessionModel, error) {
	sqlText := `select * from portal.openid_session where pk = :urn or code = :urn;`

	sqlParams := map[string]interface{}{"urn": pk}
	var sqlResults []*OpenidSessionModel

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

func UpdateOpenidSession(pk string, content string) error {
	sqlText := `update portal.openid_session set content = :content where pk = :urn or code = :urn;`

	sqlParams := map[string]interface{}{
		"urn":     pk,
		"content": content,
	}

	_, err := datastore.NamedExec(sqlText, sqlParams)
	if err != nil {
		return fmt.Errorf("UpdateOpenidSession: %w", err)
	}
	return nil
}
