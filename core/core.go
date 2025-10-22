package main

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sevlyar/go-daemon"
)

func main() {

	checkFlags()

	// Initialize logger
	consoleWriter := zerolog.ConsoleWriter{
		Out:          os.Stderr,
		TimeFormat:   "2006-01-02 15:04:05",
		TimeLocation: time.Local,
	}
	log.Logger = log.Output(consoleWriter)

	// Load .env file
	if err := godotenv.Load("../.env"); err != nil {
		if err := godotenv.Load(); err != nil {
			log.Info().Msg("No .env file found, using environment variables")
		}
	}

	// Set log level
	if os.Getenv("DOCKER_ENV") == "true" {
		setLogLevel()
	}

	// If running outside of docker, run as a daemon
	if os.Getenv("DOCKER_ENV") != "true" {

		execPath, _ := os.Executable()
		execDir := filepath.Dir(execPath)

		cntxt := &daemon.Context{
			PidFileName: filepath.Join(execDir, "skystats.pid"),
			PidFilePerm: 0644,
			LogFileName: filepath.Join(execDir, "skystats.log"),
			LogFilePerm: 0640,
			WorkDir:     execDir,
			Umask:       027,
		}

		d, err := cntxt.Reborn()

		if err != nil {
			log.Fatal().Err(err).Msg("Failed to launch daemon")
		}
		if d != nil {
			return
		}
		defer cntxt.Release()

		// when running as daemon, logs are written to file, disable ansi colors + set log level
		consoleWriter.NoColor = true
		log.Logger = log.Output(consoleWriter)
		setLogLevel()

		log.Info().Msg("Skystats: Running in daemon mode")
	}

	url := GetConnectionUrl()

	log.Info().Msg("Connecting to Postgres database")

	pg, err := NewPG(context.Background(), url)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Postgres database")
		os.Exit(1)
	}

	// Setup db
	log.Info().Msg("Checking to see if any database initialisation / migrations are needed")
	if err := RunDatabaseMigrations(); err != nil {
		log.Error().Err(err).Msg("Error initialising or migrating the database")
		os.Exit(1)
	}

	log.Info().Msg("Checking if interesting aircraft reference data needs updating from plane-alert-db")
	if err := UpsertPlaneAlertDb(pg); err != nil {
		log.Error().Msgf("Error updating interesting aircraft data: %v", err)
		os.Exit(1)
	}

	// Start API server in a separate goroutine
	log.Info().Msg("Starting API server")
	go func() {
		apiServer := NewAPIServer(pg)
		apiServer.Start()
	}()

	log.Info().Msg("Starting scheduled tasks")

	updateAircraftDataTicker := time.NewTicker(2 * time.Second)
	updateStatisticsTicker := time.NewTicker(120 * time.Second)
	updateRegistrationsTicker := time.NewTicker(30 * time.Second)
	updateRoutesTicker := time.NewTicker(300 * time.Second)
	updateInterestingSeenTicker := time.NewTicker(120 * time.Second)

	// Welcome to skystats
	if banner, err := os.ReadFile("../docs/logo/skystats_ascii.txt"); err == nil {
		log.Info().Msg("\n" + string(banner))
	}
	log.Info().Msg("Welcome to Skystats!")

	defer func() {
		log.Info().Msg("Closing database connection")
		updateAircraftDataTicker.Stop()
		updateStatisticsTicker.Stop()
		updateRegistrationsTicker.Stop()
		updateRoutesTicker.Stop()
		updateInterestingSeenTicker.Stop()
		pg.Close()
	}()

	for {
		select {
		case <-updateAircraftDataTicker.C:
			log.Debug().Msg("Update Aircraft")
			updateAircraftDatabase(pg)
		case <-updateStatisticsTicker.C:
			log.Debug().Msg("Update Statistics")
			updateMeasurementStatistics(pg)
		case <-updateRegistrationsTicker.C:
			log.Debug().Msg("Update Aircraft Registration")
			updateRegistrations(pg)
		case <-updateRoutesTicker.C:
			log.Debug().Msg("Update Routes")
			updateRoutes(pg)
		case <-updateInterestingSeenTicker.C:
			log.Debug().Msg("Update Interesting Seen")
			updateInterestingSeen(pg)
		}
	}

}

func checkFlags() {
	flag.Parse()
	if showVersion {
		showVersionExit()
	}
}

func setLogLevel() {
	switch os.Getenv("LOG_LEVEL") {
	case "DEBUG":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Info().Msg("Log level set to DEBUG")
	case "INFO":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		log.Info().Msg("Log level set to INFO")
	case "WARN":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
		log.Warn().Msg("Log level set to WARN")
	case "ERROR":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
		log.Error().Msg("Log level set to ERROR")
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		log.Info().Msg("Log level not set or invalid, defaulting to INFO")
	}
}
