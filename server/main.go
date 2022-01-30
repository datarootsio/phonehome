package main

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func main() {
	InitConfig()
	if err := InitDBConn(); err != nil {
		log.Fatal().Err(err).Msg("cannot connect to db")
	}
	g := buildServer()
	g.Run(fmt.Sprintf(":%s", viper.GetString("PORT")))
}
