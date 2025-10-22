package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

func updateRegistrations(pg *postgres) {

	aircrafts := unprocessedRegistrations(pg)

	if len(aircrafts) == 0 {
		return
	}

	existing, new := checkRegistrationExists(pg, aircrafts)

	if len(new) > 50 {
		new = new[:50]
	}

	var registrations []RegistrationInfo

	for _, aircraft := range new {

		registration, err := getRegistration(aircraft)

		if err != nil {
			log.Error().Err(err).Msg("Error fetching registration for " + aircraft.Hex)
			continue
		}

		if registration.Response.Aircraft.ModeS == "" {
			log.Debug().Msgf("No registration found for %s \n", aircraft.Hex)
			existing = append(existing, aircraft)
			continue
		}

		registrations = append(registrations, *registration)

		existing = append(existing, aircraft)

	}

	insertRegistrations(pg, registrations)

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
			log.Error().Err(err).Msg("insertRegistrations() - Unable to insert data")
		}
	}

}

func getRegistration(aircraft Aircraft) (*RegistrationInfo, error) {

	url := "https://api.adsbdb.com/v0/aircraft/"
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

func unprocessedRegistrations(pg *postgres) []Aircraft {

	query := `
		SELECT id, hex
		FROM aircraft_data
		WHERE 
			hex != '' AND
			registration_processed = false
		ORDER BY first_seen ASC`

	rows, err := pg.db.Query(context.Background(), query)

	if err != nil {
		log.Error().Err(err).Msg("unprocessedRegistrations() - Error querying db")
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
			log.Error().Err(err).Msg("unprocessedRegistrations() - Error scanning rows")
			return nil
		}

		aircrafts = append(aircrafts, aircraft)
	}

	log.Debug().Msgf("Aircrafts that have not have registration processed: %d", len(aircrafts))
	return aircrafts
}

func checkRegistrationExists(pg *postgres, aircraftToProcess []Aircraft) (existing []Aircraft, new []Aircraft) {

	var hexValues []string
	for _, a := range aircraftToProcess {
		hexValues = append(hexValues, a.Hex)
	}

	existingRegistrations := make(map[string]*Aircraft)

	query := `
		SELECT id, mode_s
		FROM registration_data
		WHERE mode_s = ANY($1::text[])`

	rows, err := pg.db.Query(context.Background(), query, hexValues)

	if err != nil {
		log.Error().Err(err).Msg("checkRegistrationExists() - Error querying db")
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
			log.Error().Err(err).Msg("checkRegistrationExists() - Error scanning rows")
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
