package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
)

func updateRoutes(pg *postgres) {

	aircrafts := unprocessed(pg)

	if len(aircrafts) == 0 {
		fmt.Println("No aircrafts to process")
		return
	}

	existing, new := checkRegistrationExists(pg, aircrafts)

	fmt.Println("Unprocessed: ", len(aircrafts))
	fmt.Println("Existing: ", len(existing))
	fmt.Println("New: ", len(new))

	if len(new) > 50 {
		new = new[:50]
	}

	var registrations []RegistrationInfo

	for _, aircraft := range new {

		registration, err := getRoute(aircraft)

		if err != nil {
			fmt.Println("Error getting route: ", err)
			continue
		}

		if registration.Response.Aircraft.ModeS == "" {
			fmt.Printf("\nNo registration found for %s", aircraft.Hex)
			continue
		}

		fmt.Printf("\nResgistration for %s: %v", aircraft.Hex, registration)

		registrations = append(registrations, *registration)

		existing = append(existing, aircraft)

	}

	insertRegistrations(pg, registrations)

	fmt.Printf("Existing about to be marked as processed: %v", existing)
	MarkProcessed(pg, "registration_processed", existing)

}

func insertRegistrations(pg *postgres, registrations []RegistrationInfo) {

	batch := &pgx.Batch{}

	for _, registration := range registrations {
		insertStatement := `
			INSERT INTO registration_data (
				type,
				icao_type,
				manufacturer,
				mode_s,
				registration,
				registered_owner_country_iso_name,
				registered_owner_country_name,
				registered_owner_operator_flag_code,
				registered_owner,
				url_photo,
				url_photo_thumbnail) 
			VALUES ( 
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			ON CONFLICT (mode_s)
			DO UPDATE SET
				type = EXCLUDED.type,
				icao_type = EXCLUDED.icao_type,
				manufacturer = EXCLUDED.manufacturer,
				registration = EXCLUDED.registration,
				registered_owner_country_iso_name = EXCLUDED.registered_owner_country_iso_name,
				registered_owner_country_name = EXCLUDED.registered_owner_country_name,
				registered_owner_operator_flag_code = EXCLUDED.registered_owner_operator_flag_code,
				registered_owner = EXCLUDED.registered_owner,
				url_photo = EXCLUDED.url_photo,
				url_photo_thumbnail = EXCLUDED.url_photo_thumbnail`

		batch.Queue(insertStatement,
			registration.Response.Aircraft.Type,
			registration.Response.Aircraft.IcaoType,
			registration.Response.Aircraft.Manufacturer,
			strings.ToLower(registration.Response.Aircraft.ModeS),
			registration.Response.Aircraft.Registration,
			registration.Response.Aircraft.RegisteredOwnerCountryIsoName,
			registration.Response.Aircraft.RegisteredOwnerCountryName,
			registration.Response.Aircraft.RegisteredOwnerOperatorFlagCode,
			registration.Response.Aircraft.RegisteredOwner,
			registration.Response.Aircraft.URLPhoto,
			registration.Response.Aircraft.URLPhotoThumbnail)
	}

	br := pg.db.SendBatch(context.Background(), batch)
	defer br.Close()

	for i := 0; i < len(registrations); i++ {
		_, err := br.Exec()
		if err != nil {
			fmt.Println("insertRegistrations() - Unable to insert data: ", err)
		}
	}

}

func getRoute(aircraft Aircraft) (*RegistrationInfo, error) {

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

	var registrationResponse RegistrationInfo
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
		ORDER BY first_seen ASC`

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
		// hexValues = append(hexValues, strings.ToUpper(a.Hex))
		hexValues = append(hexValues, a.Hex)
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
