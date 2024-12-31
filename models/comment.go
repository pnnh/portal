package models

import (
	"fmt"
	"time"

	"multiverse-authorization/neutron/services/datastore"
)

type CommentModel struct {
	Urn         string    `json:"urn"`     // 主键标识
	Title       string    `json:"title"`   // 标题
	Content     string    `json:"content"` // 内容
	CreateTime  time.Time `json:"create_time" db:"create_time"`
	UpdateTime  time.Time `json:"update_time" db:"update_time"`
	Creator     string    `json:"creator"`
	Thread      string    `json:"thread"`
	Referer     string    `json:"referer"`
	Resource    string    `json:"resource"`
	IPAddress   string    `json:"ip_address"`
	Fingerprint string    `json:"fingerprint"`
	EMail       string    `json:"email"`
	Nickname    string    `json:"nickname"`
	Website     string    `json:"website"`
}

func PGInsertComment(model *CommentModel) error {
	sqlText := `insert into comments(urn, content, create_time, update_time, creator, thread, referer, 
        resource, ipaddress, fingerprint, email, nickname, website)
values(:urn, :content, now(), now(), :creator, :thread, :referer, :resource, :ipaddress, :fingerprint, 
       :email, :nickname, :website)
on conflict (urn)
do update set content = excluded.content, update_time = now();`

	sqlParams := map[string]interface{}{
		"urn":         model.Urn,
		"content":     model.Content,
		"creator":     model.Creator,
		"thread":      model.Thread,
		"referer":     model.Referer,
		"resource":    model.Resource,
		"ipaddress":   model.IPAddress,
		"fingerprint": model.Fingerprint,
		"email":       model.EMail,
		"nickname":    model.Nickname,
		"website":     model.Website,
	}

	_, err := datastore.NamedExec(sqlText, sqlParams)
	if err != nil {
		return fmt.Errorf("PGInsertComment: %w", err)
	}
	return nil
}
