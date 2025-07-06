package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/sevlyar/go-daemon"
)

func main() {

	// Load .env file if it exists (optional for Docker)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Skip daemonization if running in Docker
	if os.Getenv("DOCKER_ENV") != "true" {
		cntxt := &daemon.Context{
			PidFileName: "skystats.pid",
			PidFilePerm: 0644,
			LogFileName: "skystats.log",
			LogFilePerm: 0640,
			WorkDir:     "./",
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
	}

	log.Print("- - - - - - - - - - - - - - -")
	log.Print("daemon started")

	url := GetConnectionUrl()
	log.Printf("Connecting to database: %s", url)

	pg, err := NewPG(context.Background(), url)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Start API server in a separate goroutine
	go func() {
		apiServer := NewAPIServer(pg)
		apiServer.Start()
	}()

	updateAircraftDataTicker := time.NewTicker(2 * time.Second)
	updateStatisticsTicker := time.NewTicker(120 * time.Second)
	updateRegistrationsTicker := time.NewTicker(60 * time.Second)
	updateRoutesTicker := time.NewTicker(60 * time.Second)
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
			fmt.Println("Update Aircraft: ", time.Now().Format("2006-01-02 15:04:05"))
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
