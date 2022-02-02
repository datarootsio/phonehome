package main

import (
	"encoding/json"
	"strconv"
	"time"

	pgd "github.com/jinzhu/gorm/dialects/postgres"
)

type BadgeInfo struct {
	SchemaVersion int64
	Label         string
	Message       string
	Color         string
}

func (bi BadgeInfo) Get(count int64) BadgeInfo {
	bi.SchemaVersion = 1
	bi.Label = "pings"
	bi.Message = strconv.FormatInt(count, 10)
	bi.Color = "brightgreen"

	return bi
}

// type DefaultResp struct {
// 	Error string      `json:"error,omitempty"`
// 	Query FilterQuery `json:"query,omitempty"`
// }

// type CallsResp struct {
// 	DefaultResp
// 	Data []Call `json:"data,omitempty"`
// }

type Call struct {
	ID           uint `gorm:"primaryKey" json:"-"`
	Timestamp    time.Time
	Payload      pgd.Jsonb `gorm:"type:jsonb" json:"payload" swaggertype:"object"`
	Organisation string    `gorm:"not null" json:"organisation"`
	Repository   string    `gorm:"not null" json:"repository"`
}

type DayCounts []struct {
	Date  time.Time `json:"date,omitempty"`
	Count int64     `json:"count,omitempty"`
}

func (dc DayCounts) MarshalJSON() ([]byte, error) {
	type dcr struct {
		Date  string `json:"date"`
		Count int64  `json:"count"`
	}

	var dcrs = []dcr{}

	const layout = "2006-01-02"
	for _, e := range dc {
		dcrs = append(dcrs, dcr{
			Date:  e.Date.Format(layout),
			Count: e.Count,
		})

	}

	return json.Marshal(&dcrs)
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
