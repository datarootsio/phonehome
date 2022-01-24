package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	pgd "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

type Call struct {
	ID           uint `gorm:"primaryKey"`
	Timestamp    time.Time
	Payload      pgd.Jsonb
	Organisation string
	Repository   string
}

func autoMigrate() error {
	if err := db.AutoMigrate(&Call{}); err != nil {
		return err
	}
	return nil
}

func registerCall(c Call) error {
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

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
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
