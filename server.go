package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Ping struct {
	ID        uint `gorm:"primaryKey"`
	Timestamp time.Time
	Origin    string
	Key       string
	Value     string
}

func registerPing() {

}

func buildServer() *gin.Engine {
	r := gin.Default()
	r.POST("/:user/:repository")
	return r
}

func InitConfig(sugar *zap.SugaredLogger) {
	viper.SetConfigFile("settings.yaml")

	if err := viper.ReadInConfig(); err != nil {
		sugar.Fatalf("cant read settings file, %v", err)
	}

	sec := viper.New()
	sec.SetConfigFile(".secrets.yaml")
	if err := sec.ReadInConfig(); err != nil {
		sugar.Infof("not process secrets file, %v", err)
	}

	viper.MergeConfig(sec)
}

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	sugar := logger.Sugar()
	InitConfig(sugar)

	sugar.Info(viper.GetString("PG_PASS"))

	// dsn := "host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai"
	// db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	// g := buildServer()
	// g.Run()

}
