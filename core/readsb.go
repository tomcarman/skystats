package main

import (
	"io"
	"net/http"
	"os"
)

func Fetch() ([]byte, error) {

	url := "http://" + os.Getenv("ADSB_HOST") + ":" + os.Getenv("ADSB_PORT") + "/data/aircraft.json"

	response, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(response.Body)

	if err != nil {
		return nil, err
	}

	return data, nil
}
