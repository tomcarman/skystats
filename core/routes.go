package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
	"github.com/tomcarman/skystats/data"
)

func getDistanceBetweenAirports(origin []float64, destination []float64) *float64 {
	distance := getRuler().Distance(origin, destination)
	return &distance
}

func updateRoutes(pg *postgres) {

	aircrafts := unprocessedRoutes(pg)

	if len(aircrafts) == 0 {
		return
	}

	if len(aircrafts) > 100 {
		aircrafts = aircrafts[:100]
	}

	existing, new := checkRouteExists(pg, aircrafts)

	routes, err := getRoutes(new)
	if err != nil {
		log.Error().Err(err).Msg("Error getting routes")
		return
	}

	insertRoutes(pg, routes)

	existing = append(existing, new...)
	MarkProcessed(pg, "route_processed", existing)

}

func unprocessedRoutes(pg *postgres) []Aircraft {

	query := `
		SELECT id, flight, last_seen_lat, last_seen_lon
		FROM aircraft_data
		WHERE 
			hex != '' AND
			flight != '' AND
			route_processed = false
		ORDER BY first_seen ASC`

	rows, err := pg.db.Query(context.Background(), query)

	if err != nil {
		log.Error().Err(err).Msg("unprocessedRoutes() - Error querying db")
		return nil
	}
	defer rows.Close()

	var aircrafts []Aircraft

	for rows.Next() {

		var aircraft Aircraft

		err := rows.Scan(
			&aircraft.Id,
			&aircraft.Flight,
			&aircraft.LastSeenLat,
			&aircraft.LastSeenLon,
		)

		if err != nil {
			log.Error().Err(err).Msg("unprocessedRoutes() - Error scanning rows")
			return nil
		}

		aircrafts = append(aircrafts, aircraft)
	}

	log.Debug().Msgf("Aircrafts that have not have routes processed: %d", len(aircrafts))
	return aircrafts
}

func checkRouteExists(pg *postgres, aircraftToProcess []Aircraft) (existing []Aircraft, new []Aircraft) {

	var callsignValues []string
	for _, a := range aircraftToProcess {
		callsignValues = append(callsignValues, a.Flight)
	}

	existingRoutes := make(map[string]*Aircraft)

	query := `
		SELECT id, route_callsign
		FROM route_data
		WHERE route_callsign = ANY($1::text[])
		  AND last_updated IS NOT NULL
		  AND last_updated > NOW() - INTERVAL '1 hour'`

	rows, err := pg.db.Query(context.Background(), query, callsignValues)

	if err != nil {
		log.Error().Err(err).Msg("checkRouteExists() - Error querying db")
		return nil, nil
	}
	defer rows.Close()

	for rows.Next() {
		var route Aircraft
		err := rows.Scan(
			&route.Id,
			&route.Flight,
		)

		if err != nil {
			log.Error().Err(err).Msg("checkRouteExists() - Error scanning rows")
			continue
		}

		existingRoutes[route.Flight] = &route
	}

	for _, a := range aircraftToProcess {
		if _, ok := existingRoutes[a.Flight]; ok {
			existing = append(existing, a)
		} else {
			new = append(new, a)
		}
	}

	return existing, new

}

func getRoutes(aircrafts []Aircraft) ([]RouteInfo, error) {

	requestBodyData := buildRouteApiRequestBody(aircrafts)
	requestBodyJson, err := json.Marshal(requestBodyData)

	if err != nil {
		return nil, err
	}

	url := "http://adsb.im/api/0/routeset"

	req, err := http.NewRequest("POST", url, bytes.NewReader(requestBodyJson))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", fmt.Sprintf("Skystats/%s", version))
	req.Header.Set("Accept", "application/json")

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var routes []RouteInfo
	err = json.Unmarshal(body, &routes)
	if err != nil {
		return nil, err
	}

	return routes, nil
}

func insertRoutes(pg *postgres, routes []RouteInfo) {

	batch := &pgx.Batch{}
	last_updated := time.Now().UTC().Format("2006-01-02 15:04:05-07")
	countryLookup := CountryIsoToName()
	queuedCount := 0

	for _, route := range routes {

		// Skip callsigns that were not matched
		if route.AirportCodesIata == "unknown" {
			continue
		}

		// Skip any "unplausible" routes
		if route.Plausible == false {
			continue
		}

		// Skip any empty or multihop routes - for now
		if route.Airports == nil || len(route.Airports) != 2 {
			continue
		}

		origin := route.Airports[0]
		destination := route.Airports[1]

		// Get country names from ISO codes
		originCountry, _ := countryLookup.GetName(origin.CountryIso2)
		destinationCountry, _ := countryLookup.GetName(destination.CountryIso2)

		// Get airline info from code
		airline, _ := data.LookupAirline(route.AirlineCode)

		// Calculate distance between airports
		var distance *float64
		if origin.Lat != 0 && origin.Lon != 0 &&
			destination.Lat != 0 && destination.Lon != 0 {
			distance = getDistanceBetweenAirports([]float64{origin.Lon, origin.Lat}, []float64{destination.Lon, destination.Lat})
		}

		insertStatement := `
			INSERT INTO route_data (
				route_callsign,
				route_callsign_icao,
				airline_name,
				airline_icao,
				airline_iata,
				origin_country_iso_name,
				origin_country_name,
				origin_elevation,
				origin_iata_code,
				origin_icao_code,
				origin_latitude,
				origin_longitude,
				origin_municipality,
				origin_name,
				destination_country_iso_name,
				destination_country_name,
				destination_elevation,
				destination_iata_code,
				destination_icao_code,
				destination_latitude,
				destination_longitude,
				destination_municipality,
				destination_name,
				last_updated,
				route_distance)
			VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
				$16, $17, $18, $19, $20, $21, $22, $23, $24, $25)
			ON CONFLICT (route_callsign)
			DO UPDATE SET
				route_callsign = EXCLUDED.route_callsign,
				route_callsign_icao = EXCLUDED.route_callsign_icao,
				airline_name = EXCLUDED.airline_name,
				airline_icao = EXCLUDED.airline_icao,
				airline_iata = EXCLUDED.airline_iata,
				origin_country_iso_name = EXCLUDED.origin_country_iso_name,
				origin_country_name = EXCLUDED.origin_country_name,
				origin_elevation = EXCLUDED.origin_elevation,
				origin_iata_code = EXCLUDED.origin_iata_code,
				origin_icao_code = EXCLUDED.origin_icao_code,
				origin_latitude = EXCLUDED.origin_latitude,
				origin_longitude = EXCLUDED.origin_longitude,
				origin_municipality = EXCLUDED.origin_municipality,
				origin_name = EXCLUDED.origin_name,
				destination_country_iso_name = EXCLUDED.destination_country_iso_name,
				destination_country_name = EXCLUDED.destination_country_name,
				destination_elevation = EXCLUDED.destination_elevation,
				destination_iata_code = EXCLUDED.destination_iata_code,
				destination_icao_code = EXCLUDED.destination_icao_code,
				destination_latitude = EXCLUDED.destination_latitude,
				destination_longitude = EXCLUDED.destination_longitude,
				destination_municipality = EXCLUDED.destination_municipality,
				destination_name = EXCLUDED.destination_name,
				last_updated = EXCLUDED.last_updated,
				route_distance = EXCLUDED.route_distance`

		batch.Queue(insertStatement,
			route.Callsign,
			route.Callsign,
			airline.Name,
			route.AirlineCode,
			airline.IATA,
			origin.CountryIso2,
			originCountry,
			origin.AltFeet,
			origin.Iata,
			origin.Icao,
			origin.Lat,
			origin.Lon,
			origin.Location,
			origin.Name,
			destination.CountryIso2,
			destinationCountry,
			destination.AltFeet,
			destination.Iata,
			destination.Icao,
			destination.Lat,
			destination.Lon,
			destination.Location,
			destination.Name,
			last_updated,
			distance)
		queuedCount++
	}

	br := pg.db.SendBatch(context.Background(), batch)
	defer br.Close()

	for i := 0; i < queuedCount; i++ {
		_, err := br.Exec()
		if err != nil {
			log.Error().Err(err).Msg("insertRoutes() - Unable to insert data")
		}
	}

}

func buildRouteApiRequestBody(aircrafts []Aircraft) RouteAPIRequest {

	aircraftsJson := make([]RouteAPIPlane, 0)

	for _, aircraft := range aircrafts {
		if aircraft.Flight != "" && aircraft.LastSeenLat.Valid && aircraft.LastSeenLon.Valid {
			aircraftsJson = append(aircraftsJson, RouteAPIPlane{
				Callsign: aircraft.Flight,
				Lat:      aircraft.LastSeenLat.Float64,
				Lng:      aircraft.LastSeenLon.Float64,
			})
		}
	}
	return RouteAPIRequest{Planes: aircraftsJson}
}
