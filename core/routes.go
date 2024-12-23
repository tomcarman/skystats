package skystats

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
)

func updateRoutes(pg *postgres) {

	aircrafts := unprocessedRoutes(pg)

	if len(aircrafts) == 0 {
		return
	}

	existing, new := checkRouteExists(pg, aircrafts)

	if len(new) > 50 {
		new = new[:50]
	}

	var routes []RouteInfo

	for _, aircraft := range new {

		route, err := getRoute(aircraft)

		if err != nil {
			fmt.Println("Error getting route: ", err)
			continue
		}

		if route.Response.Flightroute.Callsign == "" {
			// fmt.Printf("No route found for %s", aircraft.Flight)
			existing = append(existing, aircraft)
			continue
		}

		routes = append(routes, *route)
		existing = append(existing, aircraft)

	}

	insertRoutes(pg, routes)

	MarkProcessed(pg, "route_processed", existing)

}

func insertRoutes(pg *postgres, routes []RouteInfo) {

	batch := &pgx.Batch{}

	for _, route := range routes {
		insertStatement := `
			INSERT INTO route_data (
				route_callsign,
				route_callsign_icao,
				route_callsign_iata,
				airline_name,
				airline_icao,
				airline_iata,
				airline_country,
				airline_country_iso,
				airline_callsign,
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
				destination_name)
			VALUES ( 
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, 
				$15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27)
			ON CONFLICT (route_callsign)
			DO UPDATE SET
				route_callsign = EXCLUDED.route_callsign,
				route_callsign_icao = EXCLUDED.route_callsign_icao,
				route_callsign_iata = EXCLUDED.route_callsign_iata,
				airline_name = EXCLUDED.airline_name,
				airline_icao = EXCLUDED.airline_icao,
				airline_iata = EXCLUDED.airline_iata,
				airline_country = EXCLUDED.airline_country,
				airline_country_iso = EXCLUDED.airline_country_iso,
				airline_callsign = EXCLUDED.airline_callsign,
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
				destination_name = EXCLUDED.destination_name`

		batch.Queue(insertStatement,
			route.Response.Flightroute.Callsign,
			route.Response.Flightroute.CallsignIcao,
			route.Response.Flightroute.CallsignIata,
			route.Response.Flightroute.Airline.Name,
			route.Response.Flightroute.Airline.Icao,
			route.Response.Flightroute.Airline.Iata,
			route.Response.Flightroute.Airline.Country,
			route.Response.Flightroute.Airline.CountryIso,
			route.Response.Flightroute.Airline.Callsign,
			route.Response.Flightroute.Origin.CountryIsoName,
			route.Response.Flightroute.Origin.CountryName,
			route.Response.Flightroute.Origin.Elevation,
			route.Response.Flightroute.Origin.IataCode,
			route.Response.Flightroute.Origin.IcaoCode,
			route.Response.Flightroute.Origin.Latitude,
			route.Response.Flightroute.Origin.Longitude,
			route.Response.Flightroute.Origin.Municipality,
			route.Response.Flightroute.Origin.Name,
			route.Response.Flightroute.Destination.CountryIsoName,
			route.Response.Flightroute.Destination.CountryName,
			route.Response.Flightroute.Destination.Elevation,
			route.Response.Flightroute.Destination.IataCode,
			route.Response.Flightroute.Destination.IcaoCode,
			route.Response.Flightroute.Destination.Latitude,
			route.Response.Flightroute.Destination.Longitude,
			route.Response.Flightroute.Destination.Municipality,
			route.Response.Flightroute.Destination.Name)
	}

	br := pg.db.SendBatch(context.Background(), batch)
	defer br.Close()

	for i := 0; i < len(routes); i++ {
		_, err := br.Exec()
		if err != nil {
			fmt.Println("insertRoutes() - Unable to insert data: ", err)
		}
	}

}

func getRoute(aircraft Aircraft) (*RouteInfo, error) {

	url := os.Getenv("ADSB_DB_CALLSIGN_ENDPOINT")
	url += aircraft.Flight

	// fmt.Println("\nGetting route for: ", url)

	response, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(response.Body)

	if err != nil {
		return nil, err
	}

	var routeResponse RouteInfo
	json.Unmarshal(data, &routeResponse)

	return &routeResponse, nil

}

func checkRouteExists(pg *postgres, aircraftToProcess []Aircraft) (existing []Aircraft, new []Aircraft) {

	var callsignValues []string
	for _, a := range aircraftToProcess {
		callsignValues = append(callsignValues, a.Flight)
	}

	// fmt.Println("Callsigns to check: ", callsignValues)

	existingRoutes := make(map[string]*Aircraft)

	query := `
		SELECT id, route_callsign
		FROM route_data
		WHERE route_callsign = ANY($1::text[])`

	rows, err := pg.db.Query(context.Background(), query, callsignValues)

	if err != nil {
		fmt.Println("checkRouteExists() - Error querying db: ", err)
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
			fmt.Println("checkRouteExists() - Error scanning rows: ", err)
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

	// fmt.Println("Existing routes: ", len(existing))
	// fmt.Println("New routes: ", len(new))

	return existing, new

}

func unprocessedRoutes(pg *postgres) []Aircraft {

	query := `
		SELECT id, flight
		FROM aircraft_data
		WHERE 
			hex != '' AND
			flight != '' AND
			route_processed = false
		ORDER BY first_seen ASC`

	rows, err := pg.db.Query(context.Background(), query)

	if err != nil {
		fmt.Println("unprocessedRoutes() - Error querying db: ", err)
		return nil
	}
	defer rows.Close()

	var aircrafts []Aircraft

	for rows.Next() {

		var aircraft Aircraft

		err := rows.Scan(
			&aircraft.Id,
			&aircraft.Flight,
		)

		if err != nil {
			fmt.Println("unprocessedRoutes() - Error scanning rows: ", err)
			return nil
		}

		aircrafts = append(aircrafts, aircraft)
	}

	fmt.Println("Aircrafts that have not have routes processed: ", len(aircrafts))
	return aircrafts
}
