package main

import (
	"flag"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	modeFlag = flag.String("mode", "dev", "dev/test/production")
)

func parseFlag() {
	flag.Parse()
	log.Info().
		Str("version", SemVer).
		Str("gitCommit", GitCommit).
		Msg("APP starting ...")

	switch *modeFlag {
	case "dev":
		log.Info().
			Int("pid", os.Getpid()).
			Msg("pid for dlv debug attach")
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "prod":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
	log.Info().
		Str("mode", *modeFlag).
		Msg("APP arguments")
}
