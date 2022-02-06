package main

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	pgd "github.com/jinzhu/gorm/dialects/postgres"
)

const YYYYMMDDLayout = "2006-01-02"

type BadgeInfo struct {
	SchemaVersion int64
	Label         string
	Message       string
	Color         string
}

func (bi BadgeInfo) Create(count int64) BadgeInfo {
	bi.SchemaVersion = 1
	bi.Label = "telemetry"
	bi.Message = strconv.FormatInt(count, 10)
	bi.Color = "brightgreen"

	return bi
}

type DefaultResp struct {
	Error string       `json:"error,omitempty"`
	Query *FilterQuery `json:"query,omitempty"`
}

type CallsResp struct {
	DefaultResp
	Data []Call `json:"data"`
}

type CountResp struct {
	DefaultResp
	Data int64 `json:"data"`
}

type DailyCountResp struct {
	DefaultResp
	Data DayCounts `json:"data"`
}

type RegisterResp struct {
	DefaultResp
	Payload json.RawMessage `json:"payload"`
	Error   string          `json:"error,omitempty"`
	Message string          `json:"message,omitempty"`
}

type Call struct {
	ID           uint      `gorm:"primaryKey" json:"-"`
	Timestamp    time.Time `json:"timestamp" swaggerignore:"true"`
	Payload      pgd.Jsonb `gorm:"type:jsonb" json:"payload" swaggertype:"object"`
	Organisation string    `gorm:"not null" json:"organisation"`
	Repository   string    `gorm:"not null" json:"repository"`
}

type CallPayload map[string]interface{}

type DayCounts []struct {
	Date  string `json:"date,omitempty"`
	Count int64  `json:"count,omitempty"`
}

type OrgRepoURI struct {
	Organisation string `uri:"organisation" binding:"required"`
	Repository   string `uri:"repository" binding:"required"`
}

type (
	JsonDate    time.Time
	FilterQuery struct {
		GroupBy      string    `form:"group_by" json:"group_by,omitempty"`
		Key          string    `form:"key" json:"key,omitempty"`
		FromDate     *JsonDate `json:"from_date,omitempty"`
		ToDate       *JsonDate `json:"to_date,omitempty"`
		Organisation string    `json:"organisation,omitempty"`
		Repository   string    `json:"repository,omitempty"`
	}
)

func (jd *JsonDate) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	t, err := time.Parse(YYYYMMDDLayout, s)
	if err != nil {
		return err
	}
	*jd = JsonDate(t)
	return nil
}

func (fq *FilterQuery) AddOrgRepo(or OrgRepoURI) {
	fq.Organisation = or.Organisation
	fq.Repository = or.Repository
}
