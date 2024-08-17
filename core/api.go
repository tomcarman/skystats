package main

import (
	"io"
	"log"
	"net/http"
	"os"
)

func Fetch() []byte {

	response, err := http.Get(os.Getenv("ADSB_HOST_AIRCRAFT_JSON"))

	if err != nil {
		log.Fatal(err.Error())
	}

	data, err := io.ReadAll(response.Body)

	if err != nil {
		log.Fatal(err)
	}

	return data
}
