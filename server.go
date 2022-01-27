package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	pgd "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
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

func getCalls(fq FilterQuery) ([]Call, error) {
	var calls []Call
	var gq *gorm.DB

	gq = db

	if fq.Organisation == "" || fq.Repository == "" {
		return calls, errors.New("please specify organisation and repository")
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
	r.POST("/:user/:repository")
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

func main() {
	InitConfig()
	if err := InitDBConn(); err != nil {
		log.Fatal().Err(err).Msg("cannot connect to db")
	}
	// g := buildServer()
	// g.Run()
}
