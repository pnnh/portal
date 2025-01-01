//go:generate generator

package models

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"portal/neutron/server/helpers"
	"portal/neutron/services/datastore"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/jmoiron/sqlx"
)

// AccountModel 账号模型 table: accounts
type AccountModel struct {
	Urn         string    `json:"urn"`      // 主键标识
	Username    string    `json:"username"` // 账号
	Password    string    `json:"-"`        // 密码
	Photo       string    `json:"-"`        // 密码
	CreateTime  time.Time `json:"create_time" db:"create_time"`
	UpdateTime  time.Time `json:"update_time" db:"update_time"`
	Nickname    string    `json:"nickname"`
	EMail       string    `json:"email"`
	Credentials string    `json:"-"`
	Session     string    `json:"-"`
	Description string    `json:"description"`
	Status      int       `json:"status"`
	Website     string    `json:"website"`
	Fingerprint string    `json:"fingerprint"`
}

func NewAccountModel(name string, displayName string) *AccountModel {
	user := &AccountModel{
		Urn:        helpers.NewPostId(),
		Username:   name,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
		Nickname:   displayName,
	}

	return user
}

func GetAccount(pk string) (*AccountModel, error) {
	sqlText := `select * from portal.accounts where pk = :urn and status = 1;`

	sqlParams := map[string]interface{}{"urn": pk}
	var sqlResults []*AccountModel

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

func GetAccountByUsername(username string) (*AccountModel, error) {
	sqlText := `select *
	from portal.accounts where username = :username and status = 1;`

	sqlParams := map[string]interface{}{"username": username}
	var sqlResults []*AccountModel

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

func PutAccount(model *AccountModel) error {
	sqlText := `insert into portal.accounts(pk, create_time, update_time, username, password, nickname, status, session)
	values(:urn, :create_time, :update_time, :username, :password, :nickname, 1, :session)`

	sqlParams := map[string]interface{}{"urn": model.Urn, "create_time": model.CreateTime, "update_time": model.UpdateTime,
		"username": model.Username, "password": model.Password, "nickname": model.Nickname,
		"session": model.Session}

	_, err := datastore.NamedExec(sqlText, sqlParams)
	if err != nil {
		return fmt.Errorf("PutAccount: %w", err)
	}
	return nil
}

func SelectAccounts(offset int, limit int) ([]*AccountModel, error) {
	sqlText := `select * from portal.accounts offset :offset limit :limit;`

	sqlParams := map[string]interface{}{"offset": offset, "limit": limit}
	var sqlResults []*AccountModel

	rows, err := datastore.NamedQuery(sqlText, sqlParams)
	if err != nil {
		return nil, fmt.Errorf("NamedQuery: %w", err)
	}
	if err = sqlx.StructScan(rows, &sqlResults); err != nil {
		return nil, fmt.Errorf("StructScan: %w", err)
	}

	return sqlResults, nil
}

func CountAccounts() (int64, error) {
	sqlText := `select count(1) as count from portal.accounts;`

	sqlParams := map[string]interface{}{}
	var sqlResults []struct {
		Count int64 `db:"count"`
	}

	rows, err := datastore.NamedQuery(sqlText, sqlParams)
	if err != nil {
		return 0, fmt.Errorf("NamedQuery: %w", err)
	}
	if err = sqlx.StructScan(rows, &sqlResults); err != nil {
		return 0, fmt.Errorf("StructScan: %w", err)
	}
	if len(sqlResults) == 0 {
		return 0, nil
	}

	return sqlResults[0].Count, nil
}

func UpdateAccountSession(model *AccountModel, sessionData *webauthn.SessionData) error {
	sessionBytes, err := json.Marshal(sessionData)
	if err != nil {
		return fmt.Errorf("序列化sessionData出错: %s", err)
	}
	sessionText := base64.StdEncoding.EncodeToString(sessionBytes)
	model.Session = sessionText

	if model.Session == "" {
		return fmt.Errorf("session is null")
	}
	sqlText := `update portal.accounts set session = :session where pk = :urn;`

	sqlParams := map[string]interface{}{"urn": model.Urn, "session": model.Session}

	_, err = datastore.NamedExec(sqlText, sqlParams)
	if err != nil {
		return fmt.Errorf("UpdateAccountSession: %w", err)
	}
	return nil
}

func UnmarshalWebauthnSession(session string) (*webauthn.SessionData, error) {
	sessionBytes, err := base64.StdEncoding.DecodeString(session)
	if err != nil {
		return nil, fmt.Errorf("反序列化session出错: %s", err)
	}
	sessionData := &webauthn.SessionData{}
	if err := json.Unmarshal(sessionBytes, sessionData); err != nil {
		return nil, fmt.Errorf("反序列化sessionData出错: %s", err)
	}
	return sessionData, nil
}

func UpdateAccountPassword(pk string, password string) error {
	sqlText := `update portal.accounts set password = :password where pk = :urn;`

	sqlParams := map[string]interface{}{"urn": pk, "password": password}

	_, err := datastore.NamedExec(sqlText, sqlParams)
	if err != nil {
		return fmt.Errorf("UpdateAccountPassword: %w", err)
	}
	return nil
}

func EnsureAccount(model *AccountModel) error {
	sqlText := `insert into accounts(urn, username, nickname, create_time, update_time, email, website, photo, fingerprint)
values (:urn, :username, :nickname, now(), now(), :email, :website, :photo, :fingerprint)
on conflict (username) do nothing;`

	sqlParams := map[string]interface{}{
		"urn":         model.Urn,
		"username":    model.Username,
		"nickname":    model.Nickname,
		"email":       model.EMail,
		"website":     model.Website,
		"photo":       model.Photo,
		"fingerprint": model.Fingerprint,
	}

	_, err := datastore.NamedExec(sqlText, sqlParams)
	if err != nil {
		return fmt.Errorf("EnsureAccount: %w", err)
	}
	return nil
}
