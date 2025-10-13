package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"github.com/sevlyar/go-daemon"
)

func main() {

	checkFlags()

	debug := os.Getenv("DEBUG") == "true"

	// Load .env file
	if err := godotenv.Load("../.env"); err != nil {
		if err := godotenv.Load(); err != nil {
			log.Println("No .env file found, using environment variables")
		}
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
			fmt.Println("Unable to run: ", err)
			log.Fatal("Unable to run: ", err)
		}
		if d != nil {
			return
		}
		defer cntxt.Release()

		log.Print("Skystats: Running in daemon mode")
	}

	// Welcome to skystats
	if banner, err := os.ReadFile("../docs/logo/skystats_ascii.txt"); err == nil {
		log.Print("\n" + string(banner))
	}

	url := GetConnectionUrl()
	log.Printf("Connecting to postgres database...")
	pg, err := NewPG(context.Background(), url)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Setup db
	log.Println("Running database initialisation / migrations...")
	if err := RunDatabaseMigrations(); err != nil {
		log.Printf("Error initialising or migrating the database: %v", err)
		os.Exit(1)
	}

	log.Println("Updating database with plane-alert-db data...")
	if err := UpsertPlaneAlertDb(pg); err != nil {
		log.Printf("Error updating interesting aircraft data: %v", err)
		os.Exit(1)
	}

	// Start API server in a separate goroutine
	log.Println("Starting API server...")
	go func() {
		apiServer := NewAPIServer(pg)
		apiServer.Start()
	}()

	updateAircraftDataTicker := time.NewTicker(2 * time.Second)
	updateStatisticsTicker := time.NewTicker(120 * time.Second)
	updateRegistrationsTicker := time.NewTicker(30 * time.Second)
	updateRoutesTicker := time.NewTicker(300 * time.Second)
	updateInterestingSeenTicker := time.NewTicker(120 * time.Second)

	defer func() {
		fmt.Println("Closing database connection")
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
			if debug {
				fmt.Println("Update Aircraft: ", time.Now().Format("2006-01-02 15:04:05"))
			}
			updateAircraftDatabase(pg)
		case <-updateStatisticsTicker.C:
			fmt.Println("Update Statistics: ", time.Now().Format("2006-01-02 15:04:05"))
			updateMeasurementStatistics(pg)
		case <-updateRegistrationsTicker.C:
			fmt.Println("Update Registrations: ", time.Now().Format("2006-01-02 15:04:05"))
			updateRegistrations(pg)
		case <-updateRoutesTicker.C:
			fmt.Println("Update Routes: ", time.Now().Format("2006-01-02 15:04:05"))
			updateRoutes(pg)
		case <-updateInterestingSeenTicker.C:
			fmt.Println("Update Interesting Seen: ", time.Now().Format("2006-01-02 15:04:05"))
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
