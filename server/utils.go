package main

import (
	"errors"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gorm.io/datatypes"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDBConn() error {
	var err error
	var dsn string

	if viper.GetString("PG_SOCKET_DIR") == "" {
		spew.Dump(viper.AllSettings())

		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Europe/Brussels",
			viper.GetString("PG_HOST"),
			viper.GetString("PG_USER"),
			viper.GetString("PG_PASS"),
			viper.GetString("PG_DATABASE"),
			viper.GetInt("PG_PORT"))
	} else {
		dsn = fmt.Sprintf("host=%s/%s user=%s password=%s dbname=%s sslmode=disable TimeZone=Europe/Brussels",
			viper.GetString("PG_SOCKET_DIR"),
			viper.GetString("PG_INSTANCE_CONNECTION_NAME"),
			viper.GetString("PG_USER"),
			viper.GetString("PG_PASS"),
			viper.GetString("PG_DATABASE"))
	}

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
	viper.AutomaticEnv()

	viper.SetConfigName("settings")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("..")

	if err := viper.ReadInConfig(); err != nil {
		log.Warn().Err(err).Msgf("can't read settings file")
	}

	sec := viper.New()
	sec.SetConfigName(".secrets")
	sec.SetConfigType("yaml")
	sec.AddConfigPath(".")
	sec.AddConfigPath("..")
	if err := sec.ReadInConfig(); err != nil {
		log.Debug().Err(err).Msg("can't process secrets file")
	}

	viper.MergeConfigMap(sec.AllSettings())

	viper.SetDefault("PORT", 8888)
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
