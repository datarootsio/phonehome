package main

import "github.com/rs/zerolog/log"

func main() {
	InitConfig()
	if err := InitDBConn(); err != nil {
		log.Fatal().Err(err).Msg("cannot connect to db")
	}
	g := buildServer()
	g.Run(":8888")
}
