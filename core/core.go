package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/sevlyar/go-daemon"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	go func() {
		pprofUrl := getPprofHost() + ":" + getPprofPort()
		log.Println(http.ListenAndServe(pprofUrl, nil))
	}()

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

	log.Print("- - - - - - - - - - - - - - -")
	log.Print("daemon started")

	url := "postgres://" + getUser() + ":" + getPassword() + "@" + getHost() + ":" + getPort() + "/" + getDbName()

	pg, err := NewPG(context.Background(), url)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	updateAircraftDataTicker := time.NewTicker(2 * time.Second)
	updateStatisticsTicker := time.NewTicker(4 * time.Second)
	updateRoutesTicker := time.NewTicker(30 * time.Second)

	defer func() {
		fmt.Println("Closing database connection")
		updateAircraftDataTicker.Stop()
		updateStatisticsTicker.Stop()
		updateRoutesTicker.Stop()
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
		case <-updateRoutesTicker.C:
			fmt.Println("Update Routes: ", time.Now().Format("2006-01-02 15:04:05"))
			updateRoutes(pg)
		}
	}
}

func getDbName() string {
	return os.Getenv("DB_NAME")
}

func getUser() string {
	return os.Getenv("DB_USER")
}

func getPassword() string {
	return os.Getenv("DB_PASSWORD")
}

func getHost() string {
	return os.Getenv("DB_HOST")
}
func getPort() string {
	return os.Getenv("DB_PORT")
}

func getPprofHost() string {
	return os.Getenv("PPROF_HOST")
}

func getPprofPort() string {
	return os.Getenv("PPROF_PORT")
}
