package main

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// @title           phonehome.dev
// @version         1.0
// @description     KISS telemetry server for FOSS packages.

// @contact.name   phomehome.dev
// @contact.url    https://github.com/datarootsio/phonehome

// @license.name  MIT
// @license.url   https://github.com/datarootsio/phonehome/LICENSE

// @host      api.phonehome.dev

// @securityDefinitions.basic  BasicAuth
func main() {
	InitConfig()
	if err := InitDBConn(); err != nil {
		log.Fatal().Err(err).Msg("cannot connect to db")
	}
	g := buildServer()
	g.Run(fmt.Sprintf(":%s", viper.GetString("PORT")))
}
