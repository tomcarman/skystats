package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func updateRoutes(pg *postgres) {

	aircrafts := unprocessed(pg)

	if len(aircrafts) == 0 {
		fmt.Println("No aircrafts to process")
		return
	}

	existing, new := checkRegistrationExists(pg, aircrafts)

	fmt.Println("Existing: ", len(existing))
	fmt.Println("New: ", len(new))

	registration, err := getRoute(new[0])

	if err != nil {
		fmt.Println("Error getting route: ", err)
		return
	}

	fmt.Println("Registration: ", registration)

	// MarkProcessed(pg, "registration_processed", existing)

}

func getRoute(aircraft Aircraft) (*AdsbdbRegistration, error) {

	url := os.Getenv("ADSB_DB_AIRCRAFT_ENDPOINT")
	url += aircraft.Hex

	response, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(response.Body)

	if err != nil {
		return nil, err
	}

	var registrationResponse AdsbdbRegistration
	json.Unmarshal(data, &registrationResponse)

	return &registrationResponse, nil

}

func unprocessed(pg *postgres) []Aircraft {

	query := `
		SELECT id, hex
		FROM aircraft_data
		WHERE 
			hex != '' AND
			registration_processed = false
		ORDER BY first_seen ASC
		LIMIT 1`

	rows, err := pg.db.Query(context.Background(), query)

	if err != nil {
		fmt.Println("getAircraftWithoutRegistrationProcessed() - Error querying db: ", err)
		return nil
	}
	defer rows.Close()

	var aircrafts []Aircraft

	for rows.Next() {

		var aircraft Aircraft

		err := rows.Scan(
			&aircraft.Id,
			&aircraft.Hex,
		)

		if err != nil {
			fmt.Println("getAircraftWithoutRegistrationProcessed() - Error scanning rows: ", err)
			return nil
		}

		aircrafts = append(aircrafts, aircraft)
	}

	fmt.Println("Aircrafts that have not have route processed: ", len(aircrafts))
	return aircrafts
}

func checkRegistrationExists(pg *postgres, aircraftToProcess []Aircraft) (existing []Aircraft, new []Aircraft) {

	var hexValues []string
	for _, a := range aircraftToProcess {
		hexValues = append(hexValues, strings.ToUpper(a.Hex))
	}

	existingRegistrations := make(map[string]*Aircraft)

	query := `
		SELECT id, mode_s
		FROM registration_data
		WHERE mode_s = ANY($1::text[])`

	rows, err := pg.db.Query(context.Background(), query, hexValues)

	if err != nil {
		fmt.Println("checkIfRegistrationInformationExists() - Error querying db: ", err)
		return nil, nil
	}
	defer rows.Close()

	for rows.Next() {
		var registration Aircraft
		err := rows.Scan(
			&registration.Id,
			&registration.Hex,
		)

		if err != nil {
			fmt.Println("checkIfRegistrationInformationExists() - Error scanning rows: ", err)
			continue
		}

		existingRegistrations[registration.Hex] = &registration
	}

	for _, a := range aircraftToProcess {
		if _, ok := existingRegistrations[a.Hex]; ok {
			existing = append(existing, a)
		} else {
			new = append(new, a)
		}
	}

	return existing, new

}
