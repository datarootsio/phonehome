package main

import (
	"strconv"
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
	bi.Label = "pings"
	bi.Message = strconv.FormatInt(count, 10)
	bi.Color = "brightgreen"

	return bi
}

type DefaultResp struct {
	Error string      `json:"error,omitempty"`
	Query FilterQuery `json:"query,omitempty"`
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
	Error string `json:"error,omitempty"`
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

type FilterQuery struct {
	GroupBy      string `form:"group_by" json:"group_by,omitempty"`
	Key          string `form:"key" json:"key,omitempty"`
	FromDate     string `json:"from_date,omitempty"`
	ToDate       string `json:"to_date,omitempty"`
	Organisation string `json:"organisation,omitempty"`
	Repository   string `json:"repository,omitempty"`
}

func (fq *FilterQuery) AddOrgRepo(or OrgRepoURI) {
	fq.Organisation = or.Organisation
	fq.Repository = or.Repository
}
