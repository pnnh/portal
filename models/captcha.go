package models

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"portal/neutron/services/datastore"
)

type CaptchaModel struct {
	Pk         string    `json:"urn"`
	Content    string    `json:"content"`
	CreateTime time.Time `json:"create_time" db:"create_time"`
	UpdateTime time.Time `json:"update_time" db:"update_time"`
	Checked    int       `json:"checked" db:"checked"`
	Used       int       `json:"used" db:"used"`
}

func PutCaptcha(model *CaptchaModel) error {
	sqlText := `insert into portal.captcha(pk, content, create_time, checked, update_time,used)
	values(:urn, :content, :create_time, :checked, :update_time,:used)`

	sqlParams := map[string]interface{}{"urn": model.Pk,
		"content": model.Content, "create_time": model.CreateTime,
		"update_time": model.UpdateTime,
		"checked":     model.Checked,
		"used":        model.Used}

	_, err := datastore.NamedExec(sqlText, sqlParams)
	if err != nil {
		return fmt.Errorf("PutCaptcha: %w", err)
	}
	return nil

}

func FindCaptcha(key string) (*CaptchaModel, error) {
	sqlText := `select * from portal.captcha where pk = :urn;`

	sqlParams := map[string]interface{}{"urn": key}
	var sqlResults []*CaptchaModel

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

func UpdateCaptcha(key string, checked int) error {
	sqlText := `update portal.captcha set checked = :checked, update_time = :update_time 
	where pk = :urn;`

	sqlParams := map[string]interface{}{
		"update_time": time.Now(),
		"urn":         key,
		"checked":     checked,
	}

	_, err := datastore.NamedExec(sqlText, sqlParams)
	if err != nil {
		return fmt.Errorf("UpdateAccountPassword: %w", err)
	}
	return nil
}

func UpdateCaptchaUsed(key string, used int) error {
	sqlText := `update portal.captcha set used = :used, update_time = :update_time 
	where pk = :urn;`

	sqlParams := map[string]interface{}{
		"update_time": time.Now(),
		"urn":         key,
		"used":        used,
	}

	_, err := datastore.NamedExec(sqlText, sqlParams)
	if err != nil {
		return fmt.Errorf("UpdateAccountPassword: %w", err)
	}
	return nil
}
