package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func Fetch() []byte {

	url := "http://" + os.Getenv("ADSB_HOST") + ":" + os.Getenv("ADSB_PORT") + "/data/aircraft.json"

	response, err := http.Get(url)

	if err != nil {
		fmt.Println("Error fetching data from ", url)
		log.Fatal(err.Error())
	}

	data, err := io.ReadAll(response.Body)

	if err != nil {
		log.Fatal(err)
	}

	return data
}
