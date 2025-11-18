package main

import (
	_ "embed"
	"os"
	"strings"

	"github.com/densify-dev/gcp-oauth2-token/pkg/gcp"
	"github.com/rs/zerolog"
)

//go:generate sh -c "printf %s $(git describe --abbrev=0 --tags) > version.txt"
//go:embed version.txt
var Version string

func main() {
	logger := initLogger()
	logger.Info().Msgf("Running GCP OAuth2 Token Generator, version: %s", strings.TrimSpace(Version))
	if err := gcp.CreateTokenFile(logger); err == nil {
		logger.Info().Msg("Token file creation succeeded")
	} else {
		logger.Error().Err(err).Msg("Token file creation failed")
		os.Exit(1)
	}
}

const (
	RFC3339Micro = "2006-01-02T15:04:05.999999Z07:00"
	debugEnvVar  = "DEBUG"
)

func initLogger() zerolog.Logger {
	zerolog.TimeFieldFormat = RFC3339Micro
	var level zerolog.Level
	if _, debug := os.LookupEnv(debugEnvVar); debug {
		level = zerolog.DebugLevel
	} else {
		level = zerolog.InfoLevel
	}
	return zerolog.New(os.Stdout).Level(level).With().Timestamp().Caller().Stack().Logger()
}
