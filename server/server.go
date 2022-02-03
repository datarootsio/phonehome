package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

var db *gorm.DB

const getCallsLimit = 3000

func buildServer() *gin.Engine {
	r := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:8080", "https://phonehome.dev"}
	r.Use(cors.New(config))

	r.GET("/:organisation/:repository", getCallsHandler)
	r.GET("/:organisation/:repository/count", getCountCallsHandler)
	r.GET("/:organisation/:repository/count/daily", getCountCallsByDayHandler)
	r.GET("/:organisation/:repository/count/badge", getCountCallsBadgeHandler)

	r.POST("/:organisation/:repository", registerCallHander)

	r.StaticFile("/docs/swagger.json", "./docs/swagger.json")

	docsUrl := ginSwagger.URL("/docs/swagger.json")
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler, docsUrl))
	return r
}
