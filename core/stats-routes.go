package main

import (
	"context"
	"fmt"
	"math"
)

type RouteStats struct {
	TotalRoutes            int                    `json:"total_routes"`
	TopAirlines            []AirlineCount         `json:"top_airlines"`
	TopRoutes              []RouteCount           `json:"top_routes"`
	TopOriginAirports      []AirportCount         `json:"top_origin_airports"`
	TopDestinationAirports []AirportCount         `json:"top_destination_airports"`
	TopCountries           []CountryCount         `json:"top_countries"`
	InternationalVsDomestic DomesticIntlCount     `json:"international_vs_domestic"`
	AverageRouteDistance   float64               `json:"average_route_distance"`
}

type AirlineCount struct {
	AirlineName string `json:"airline_name"`
	AirlineICAO string `json:"airline_icao"`
	AirlineIATA string `json:"airline_iata"`
	Count       int    `json:"count"`
}

type RouteCount struct {
	Route       string `json:"route"`
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
	Count       int    `json:"count"`
}

type AirportCount struct {
	AirportCode string `json:"airport_code"`
	AirportName string `json:"airport_name"`
	Country     string `json:"country"`
	Count       int    `json:"count"`
}

type CountryCount struct {
	Country    string `json:"country"`
	CountryISO string `json:"country_iso"`
	Count      int    `json:"count"`
}

type DomesticIntlCount struct {
	Domestic      int `json:"domestic"`
	International int `json:"international"`
}

func getRouteStatistics(pg *postgres) (*RouteStats, error) {
	stats := &RouteStats{}

	// Get total routes count (point-to-point only)
	totalRoutes, err := getTotalRoutesCount(pg)
	if err != nil {
		return nil, fmt.Errorf("error getting total routes count: %v", err)
	}
	stats.TotalRoutes = totalRoutes
	
	// If no routes, return empty stats
	if totalRoutes == 0 {
		return stats, nil
	}

	// Get top airlines
	topAirlines, err := getTopAirlines(pg, 10)
	if err != nil {
		return nil, fmt.Errorf("error getting top airlines: %v", err)
	}
	stats.TopAirlines = topAirlines

	// Get top routes
	topRoutes, err := getTopRoutes(pg, 10)
	if err != nil {
		return nil, fmt.Errorf("error getting top routes: %v", err)
	}
	stats.TopRoutes = topRoutes

	// Get top origin airports
	topOriginAirports, err := getTopOriginAirports(pg, 10)
	if err != nil {
		return nil, fmt.Errorf("error getting top origin airports: %v", err)
	}
	stats.TopOriginAirports = topOriginAirports

	// Get top destination airports
	topDestinationAirports, err := getTopDestinationAirports(pg, 10)
	if err != nil {
		return nil, fmt.Errorf("error getting top destination airports: %v", err)
	}
	stats.TopDestinationAirports = topDestinationAirports

	// Get top countries
	topCountries, err := getTopCountries(pg, 10)
	if err != nil {
		return nil, fmt.Errorf("error getting top countries: %v", err)
	}
	stats.TopCountries = topCountries

	// Get international vs domestic ratio
	domesticIntl, err := getDomesticVsInternational(pg)
	if err != nil {
		return nil, fmt.Errorf("error getting domestic vs international: %v", err)
	}
	stats.InternationalVsDomestic = domesticIntl

	// Get average route distance
	avgDistance, err := getAverageRouteDistance(pg)
	if err != nil {
		return nil, fmt.Errorf("error getting average route distance: %v", err)
	}
	stats.AverageRouteDistance = avgDistance

	return stats, nil
}

func getTotalRoutesCount(pg *postgres) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM aircraft_data ad 
		INNER JOIN route_data rd ON ad.flight = rd.route_callsign
		WHERE rd.origin_iata_code != rd.destination_iata_code
			AND rd.origin_iata_code IS NOT NULL AND rd.origin_iata_code != ''
			AND rd.destination_iata_code IS NOT NULL AND rd.destination_iata_code != ''`
	
	var count int
	err := pg.db.QueryRow(context.Background(), query).Scan(&count)
	if err != nil {
		return 0, err
	}
	
	return count, nil
}

func getTopAirlines(pg *postgres, limit int) ([]AirlineCount, error) {
	query := `
		SELECT 
			rd.airline_name,
			rd.airline_icao,
			rd.airline_iata,
			COUNT(*) as flight_count
		FROM aircraft_data ad 
		INNER JOIN route_data rd ON ad.flight = rd.route_callsign
		WHERE rd.airline_name IS NOT NULL AND rd.airline_name != ''
			AND rd.origin_iata_code != rd.destination_iata_code
			AND rd.origin_iata_code IS NOT NULL AND rd.origin_iata_code != ''
			AND rd.destination_iata_code IS NOT NULL AND rd.destination_iata_code != ''
		GROUP BY rd.airline_name, rd.airline_icao, rd.airline_iata
		ORDER BY flight_count DESC
		LIMIT $1`

	rows, err := pg.db.Query(context.Background(), query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var airlines []AirlineCount
	for rows.Next() {
		var airline AirlineCount
		err := rows.Scan(&airline.AirlineName, &airline.AirlineICAO, &airline.AirlineIATA, &airline.Count)
		if err != nil {
			return nil, err
		}
		airlines = append(airlines, airline)
	}

	return airlines, nil
}

func getTopRoutes(pg *postgres, limit int) ([]RouteCount, error) {
	query := `
		SELECT 
			CONCAT(rd.origin_iata_code, ' → ', rd.destination_iata_code) as route,
			rd.origin_iata_code,
			rd.destination_iata_code,
			COUNT(*) as flight_count
		FROM aircraft_data ad 
		INNER JOIN route_data rd ON ad.flight = rd.route_callsign
		WHERE rd.origin_iata_code IS NOT NULL AND rd.origin_iata_code != ''
			AND rd.destination_iata_code IS NOT NULL AND rd.destination_iata_code != ''
			AND rd.origin_iata_code != rd.destination_iata_code
		GROUP BY rd.origin_iata_code, rd.destination_iata_code
		ORDER BY flight_count DESC
		LIMIT $1`

	rows, err := pg.db.Query(context.Background(), query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routes []RouteCount
	for rows.Next() {
		var route RouteCount
		err := rows.Scan(&route.Route, &route.Origin, &route.Destination, &route.Count)
		if err != nil {
			return nil, err
		}
		routes = append(routes, route)
	}

	return routes, nil
}

func getTopOriginAirports(pg *postgres, limit int) ([]AirportCount, error) {
	query := `
		SELECT 
			rd.origin_iata_code,
			rd.origin_name,
			rd.origin_country_name,
			COUNT(*) as flight_count
		FROM aircraft_data ad 
		INNER JOIN route_data rd ON ad.flight = rd.route_callsign
		WHERE rd.origin_iata_code IS NOT NULL AND rd.origin_iata_code != ''
			AND rd.origin_iata_code != rd.destination_iata_code
			AND rd.destination_iata_code IS NOT NULL AND rd.destination_iata_code != ''
		GROUP BY rd.origin_iata_code, rd.origin_name, rd.origin_country_name
		ORDER BY flight_count DESC
		LIMIT $1`

	rows, err := pg.db.Query(context.Background(), query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var airports []AirportCount
	for rows.Next() {
		var airport AirportCount
		err := rows.Scan(&airport.AirportCode, &airport.AirportName, &airport.Country, &airport.Count)
		if err != nil {
			return nil, err
		}
		airports = append(airports, airport)
	}

	return airports, nil
}

func getTopDestinationAirports(pg *postgres, limit int) ([]AirportCount, error) {
	query := `
		SELECT 
			rd.destination_iata_code,
			rd.destination_name,
			rd.destination_country_name,
			COUNT(*) as flight_count
		FROM aircraft_data ad 
		INNER JOIN route_data rd ON ad.flight = rd.route_callsign
		WHERE rd.destination_iata_code IS NOT NULL AND rd.destination_iata_code != ''
			AND rd.origin_iata_code != rd.destination_iata_code
			AND rd.origin_iata_code IS NOT NULL AND rd.origin_iata_code != ''
		GROUP BY rd.destination_iata_code, rd.destination_name, rd.destination_country_name
		ORDER BY flight_count DESC
		LIMIT $1`

	rows, err := pg.db.Query(context.Background(), query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var airports []AirportCount
	for rows.Next() {
		var airport AirportCount
		err := rows.Scan(&airport.AirportCode, &airport.AirportName, &airport.Country, &airport.Count)
		if err != nil {
			return nil, err
		}
		airports = append(airports, airport)
	}

	return airports, nil
}

func getTopCountries(pg *postgres, limit int) ([]CountryCount, error) {
	query := `
		SELECT 
			rd.airline_country,
			rd.airline_country_iso,
			COUNT(*) as flight_count
		FROM aircraft_data ad 
		INNER JOIN route_data rd ON ad.flight = rd.route_callsign
		WHERE rd.airline_country IS NOT NULL AND rd.airline_country != ''
			AND rd.origin_iata_code != rd.destination_iata_code
			AND rd.origin_iata_code IS NOT NULL AND rd.origin_iata_code != ''
			AND rd.destination_iata_code IS NOT NULL AND rd.destination_iata_code != ''
		GROUP BY rd.airline_country, rd.airline_country_iso
		ORDER BY flight_count DESC
		LIMIT $1`

	rows, err := pg.db.Query(context.Background(), query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var countries []CountryCount
	for rows.Next() {
		var country CountryCount
		err := rows.Scan(&country.Country, &country.CountryISO, &country.Count)
		if err != nil {
			return nil, err
		}
		countries = append(countries, country)
	}

	return countries, nil
}

func getDomesticVsInternational(pg *postgres) (DomesticIntlCount, error) {
	query := `
		SELECT 
			SUM(CASE WHEN rd.origin_country_iso_name = rd.destination_country_iso_name THEN 1 ELSE 0 END) as domestic,
			SUM(CASE WHEN rd.origin_country_iso_name != rd.destination_country_iso_name THEN 1 ELSE 0 END) as international
		FROM aircraft_data ad 
		INNER JOIN route_data rd ON ad.flight = rd.route_callsign
		WHERE rd.origin_country_iso_name IS NOT NULL AND rd.destination_country_iso_name IS NOT NULL
			AND rd.origin_iata_code != rd.destination_iata_code
			AND rd.origin_iata_code IS NOT NULL AND rd.origin_iata_code != ''
			AND rd.destination_iata_code IS NOT NULL AND rd.destination_iata_code != ''`

	var result DomesticIntlCount
	err := pg.db.QueryRow(context.Background(), query).Scan(&result.Domestic, &result.International)
	if err != nil {
		return DomesticIntlCount{}, err
	}

	return result, nil
}

func getAverageRouteDistance(pg *postgres) (float64, error) {
	query := `
		SELECT 
			AVG(distance_km) as avg_distance
		FROM (
			SELECT 
				6371 * acos(
					LEAST(1.0, GREATEST(-1.0,
						cos(radians(rd.origin_latitude)) * cos(radians(rd.destination_latitude)) * 
						cos(radians(rd.destination_longitude) - radians(rd.origin_longitude)) + 
						sin(radians(rd.origin_latitude)) * sin(radians(rd.destination_latitude))
					))
				) as distance_km
			FROM aircraft_data ad 
			INNER JOIN route_data rd ON ad.flight = rd.route_callsign
			WHERE rd.origin_latitude IS NOT NULL AND rd.origin_longitude IS NOT NULL
				AND rd.destination_latitude IS NOT NULL AND rd.destination_longitude IS NOT NULL
				AND rd.origin_latitude != 0 AND rd.origin_longitude != 0
				AND rd.destination_latitude != 0 AND rd.destination_longitude != 0
				AND rd.origin_latitude BETWEEN -90 AND 90
				AND rd.destination_latitude BETWEEN -90 AND 90
				AND rd.origin_longitude BETWEEN -180 AND 180
				AND rd.destination_longitude BETWEEN -180 AND 180
				AND rd.origin_iata_code != rd.destination_iata_code
				AND rd.origin_iata_code IS NOT NULL AND rd.origin_iata_code != ''
				AND rd.destination_iata_code IS NOT NULL AND rd.destination_iata_code != ''
		) distances`

	var avgDistance float64
	err := pg.db.QueryRow(context.Background(), query).Scan(&avgDistance)
	if err != nil {
		return 0, err
	}

	if math.IsNaN(avgDistance) {
		return 0, nil
	}

	return avgDistance, nil
}