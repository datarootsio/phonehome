package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	pgd "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/datatypes"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

const getCallsLimit = 3000

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

type Call struct {
	ID           uint `gorm:"primaryKey" json:"-"`
	Timestamp    time.Time
	Payload      pgd.Jsonb `gorm:"type:jsonb" json:"payload"`
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

func autoMigrate() error {
	if err := db.AutoMigrate(&Call{}); err != nil {
		return err
	}
	return nil
}

func callsQueryBuilder(fq FilterQuery) (*gorm.DB, error) {
	var gq *gorm.DB

	gq = db

	if fq.Organisation == "" || fq.Repository == "" {
		return nil, errors.New("please specify organisation and repository")
	}
	gq = gq.Where("organisation = ? AND repository = ?", fq.Organisation, fq.Repository)

	if fq.Key != "" {
		gq = gq.Where(datatypes.JSONQuery("payload").HasKey(fq.Key))
	}

	if fq.FromDate != "" {
		gq = gq.Where("from_date >= ?", fq.FromDate)
	}

	if fq.ToDate != "" {
		gq = gq.Where("to_date < ?", fq.ToDate)
	}

	return gq, nil
}

func getCountCalls(fq FilterQuery) (int64, error) {
	var count int64

	gq, err := callsQueryBuilder(fq)
	if err != nil {
		return count, nil
	}

	result := gq.Model(&Call{}).Count(&count)
	if result.Error != nil {
		return count, result.Error
	}

	return count, nil
}

func getCountCallsByDate(fq FilterQuery) (DayCounts, error) {
	dc := DayCounts{}

	res := db.Raw(`
	select timestamp::date as date, count(*) 
	from calls
	group by timestamp::date
	order by timestamp::date ASC
	`).Scan(&dc)
	if res.Error != nil {
		return dc, res.Error
	}

	return dc, nil
}

func getCountCallsBadgeHandler(c *gin.Context) {
	var or OrgRepoURI
	if err := c.ShouldBindUri(&or); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return

	}

	fq := FilterQuery{}
	fq.AddOrgRepo(or)

	count, err := getCountCalls(fq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	bi := BadgeInfo{}.Get(count)

	c.JSON(200, bi)

}

func getCountCallsHandler(c *gin.Context) {
	var fq FilterQuery
	var or OrgRepoURI
	resp := gin.H{}

	c.ShouldBind(&fq)
	if err := c.ShouldBindUri(&or); err != nil {
		resp["error"] = err.Error()
		c.JSON(http.StatusBadRequest, resp)
		return

	}

	fq.Organisation = or.Organisation
	fq.Repository = or.Repository
	resp["query"] = fq

	switch fq.GroupBy {
	case "":
		count, err := getCountCalls(fq)
		if err != nil {
			resp["error"] = err.Error()
			c.JSON(http.StatusBadRequest, resp)
			return
		}
		resp["data"] = count
		c.JSON(200, resp)
	case "day":
		spew.Dump("day")
		dc, err := getCountCallsByDate(fq)
		if err != nil {
			resp["error"] = err.Error()
			c.JSON(http.StatusBadRequest, resp)
			return
		}
		resp["data"] = dc
		c.JSON(200, resp)
	default:
		resp["error"] = fmt.Sprintf(`group_by key '%s' not supported`, fq.GroupBy)
		c.JSON(http.StatusBadRequest, resp)
	}
}

func getCalls(fq FilterQuery) ([]Call, error) {
	var calls []Call

	gq, err := callsQueryBuilder(fq)
	if err != nil {
		return calls, err
	}

	r := gq.Find(&calls).Limit(getCallsLimit)
	return calls, r.Error
}

func getCallsHandler(c *gin.Context) {
	var fq FilterQuery
	c.ShouldBind(&fq)
	c.ShouldBindUri(&fq)

	cs, err := getCalls(fq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"query": fq,
		"data":  cs,
	})
}

func registerCall(c Call) error {
	if !json.Valid(c.Payload.RawMessage) {
		return fmt.Errorf("%s is invalid JSON", c.Payload.RawMessage)
	}

	c.Timestamp = time.Now()

	result := db.Create(&c)
	return result.Error
}

func buildServer() *gin.Engine {
	r := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:8080"}
	r.Use(cors.New(config))

	r.GET("/:organisation/:repository", getCallsHandler)
	r.GET("/:organisation/:repository/count", getCountCallsHandler)
	r.GET("/openapi/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	r.GET("/:organisation/:repository/count/badge", getCountCallsBadgeHandler)
	return r
}

func InitDBConn() error {
	var err error
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Europe/Brussels",
		viper.GetString("PG_HOST"),
		viper.GetString("PG_USER"),
		viper.GetString("PG_PASS"),
		viper.GetString("PG_DATABASE"),
		viper.GetInt("PG_PORT"))

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
	if err != nil {
		return err
	}

	if err := autoMigrate(); err != nil {
		return err
	}

	return nil
}

func InitConfig() {
	viper.SetConfigFile("settings.yaml")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal().Err(err).Msgf("cant read settings file")
	}

	sec := viper.New()
	sec.SetConfigFile(".secrets.yaml")
	if err := sec.ReadInConfig(); err != nil {
		log.Debug().Msgf("not process secrets file, %v", err)
	}

	viper.MergeConfigMap(sec.AllSettings())

	viper.SetDefault("PORT", 8888)
}
