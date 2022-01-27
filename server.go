package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	docs "github.com/datarootsio/phonehome/docs"
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

type Call struct {
	ID           uint `gorm:"primaryKey"`
	Timestamp    time.Time
	Payload      pgd.Jsonb `gorm:"type:jsonb"`
	Organisation string    `gorm:"not null"`
	Repository   string    `gorm:"not null"`
}

type CallCount struct {
	Count int64
	Query FilterQuery
}

type FilterQuery struct {
	Key          string
	FromDate     string
	ToDate       string
	Organisation string
	Repository   string
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

func countCalls(fq FilterQuery) (CallCount, error) {
	var count int64
	var cc CallCount

	gq, err := callsQueryBuilder(fq)
	if err != nil {
		return cc, nil
	}

	result := gq.Model(&Call{}).Count(&count)
	if result.Error != nil {
		return cc, result.Error
	}
	return CallCount{
		Query: fq,
		Count: count,
	}, nil

}

func getCalls(fq FilterQuery) ([]Call, error) {
	var calls []Call

	gq, err := callsQueryBuilder(fq)
	if err != nil {
		return calls, err
	}

	r := gq.Find(&calls)
	return calls, r.Error
}

func registerCall(c Call) error {
	if !json.Valid(c.Payload.RawMessage) {
		return fmt.Errorf("%s is invalid JSON", c.Payload.RawMessage)
	}
	result := db.Create(&c)
	return result.Error
}

func buildServer() *gin.Engine {
	r := gin.Default()
	docs.SwaggerInfo.BasePath = "/api/v1"
	v1 := r.Group("/api/v1")
	v1.POST("/:user/:repository")
	v1.GET("/:user/:repository/count")
	v1.GET("/openapi/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
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
}
