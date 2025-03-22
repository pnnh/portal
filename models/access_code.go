package models

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"portal/quark/neutron/services/datastore"
)

type AccessCodeModel struct {
	Pk         string    `json:"uid"`
	CreateTime time.Time `json:"create_time" db:"create_time"`
	UpdateTime time.Time `json:"update_time" db:"update_time"`
	Code       string    `json:"code"`
	Content    string    `json:"content"`
	Active     int       `json:"active"`
}

func PutAccessCode(model *AccessCodeModel) error {
	sqlText := `insert into access_code(pk, create_time, update_time, code, content, active)
		values(:uid, :create_time, :update_time, :code, :content, :active);`

	sqlParams := map[string]interface{}{
		"uid":         model.Pk,
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
	sqlText := `delete from access_code where pk = :uid or code = :uid;`

	sqlParams := map[string]interface{}{
		"uid": pk,
	}

	_, err := datastore.NamedExec(sqlText, sqlParams)
	if err != nil {
		return fmt.Errorf("DeleteAccessCode: %w", err)
	}
	return nil
}

func GetAccessCode(pk string) (*AccessCodeModel, error) {
	sqlText := `select * from access_code where pk = :uid or code = :uid;`

	sqlParams := map[string]interface{}{"uid": pk}
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
	sqlText := `update access_code set active = :active where pk = :uid or code = :uid;`

	sqlParams := map[string]interface{}{
		"uid":    pk,
		"active": active,
	}

	_, err := datastore.NamedExec(sqlText, sqlParams)
	if err != nil {
		return fmt.Errorf("UpdateAccessCodeStatus: %w", err)
	}
	return nil
}
