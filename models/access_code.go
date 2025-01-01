package models

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"portal/neutron/services/datastore"
)

type AccessCodeModel struct {
	Pk         string    `json:"urn"`
	CreateTime time.Time `json:"create_time" db:"create_time"`
	UpdateTime time.Time `json:"update_time" db:"update_time"`
	Code       string    `json:"code"`
	Content    string    `json:"content"`
	Active     int       `json:"active"`
}

func PutAccessCode(model *AccessCodeModel) error {
	sqlText := `insert into portal.access_code(pk, create_time, update_time, code, content, active)
		values(:urn, :create_time, :update_time, :code, :content, :active);`

	sqlParams := map[string]interface{}{
		"urn":         model.Pk,
		"create_time": model.CreateTime,
		"update_time": model.UpdateTime,
		"code":        model.Code,
		"content":     model.Content,
		"active":      model.Active,
	}

	_, err := datastore.NamedExec(sqlText, sqlParams)
	if err != nil {
		return fmt.Errorf("PutSession: %w", err)
	}
	return nil
}

func DeleteAccessCode(pk string) error {
	sqlText := `delete from portal.access_code where pk = :urn or code = :urn;`

	sqlParams := map[string]interface{}{
		"urn": pk,
	}

	_, err := datastore.NamedExec(sqlText, sqlParams)
	if err != nil {
		return fmt.Errorf("DeleteAccessCode: %w", err)
	}
	return nil
}

func GetAccessCode(pk string) (*AccessCodeModel, error) {
	sqlText := `select * from portal.access_code where pk = :urn or code = :urn;`

	sqlParams := map[string]interface{}{"urn": pk}
	var sqlResults []*AccessCodeModel

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

func UpdateAccessCodeStatus(pk string, active int) error {
	sqlText := `update portal.access_code set active = :active where pk = :urn or code = :urn;`

	sqlParams := map[string]interface{}{
		"urn":    pk,
		"active": active,
	}

	_, err := datastore.NamedExec(sqlText, sqlParams)
	if err != nil {
		return fmt.Errorf("UpdateAccessCodeStatus: %w", err)
	}
	return nil
}
