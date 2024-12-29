package main

import (
	"context"
)

func getHighestAircraftFloor(pg *postgres) int {

	var returnValue int
	defaultValue := 0

	query := `SELECT barometric_altitude 
				FROM highest_aircraft 
				ORDER BY barometric_altitude ASC, first_seen ASC
				LIMIT 1`

	err := pg.db.QueryRow(context.Background(), query).Scan(&returnValue)
	if err == nil {
		return returnValue
	} else {
		return defaultValue
	}
}

func getLowestAircraftCeiling(pg *postgres) int {

	var returnValue int
	defaultValue := 999999

	query := `SELECT barometric_altitude 
				FROM lowest_aircraft 
				ORDER BY barometric_altitude DESC, first_seen ASC
				LIMIT 1`

	err := pg.db.QueryRow(context.Background(), query).Scan(&returnValue)
	if err == nil {
		return returnValue
	} else {
		return defaultValue
	}
}

func getFastestAircraftFloor(pg *postgres) float64 {

	var returnValue float64
	defaultValue := 0.0

	query := `SELECT ground_speed 
				FROM fastest_aircraft 
				ORDER BY ground_speed ASC, first_seen ASC
				LIMIT 1`

	err := pg.db.QueryRow(context.Background(), query).Scan(&returnValue)
	if err == nil {
		return returnValue
	} else {
		return defaultValue
	}
}

func getSlowestAircraftCeiling(pg *postgres) float64 {

	var returnValue float64
	defaultValue := 99999.0

	query := `SELECT ground_speed 
				FROM slowest_aircraft 
				ORDER BY ground_speed DESC, first_seen ASC
				LIMIT 1`

	err := pg.db.QueryRow(context.Background(), query).Scan(&returnValue)
	if err == nil {
		return returnValue
	} else {
		return defaultValue
	}
}
