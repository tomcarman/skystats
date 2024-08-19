package main

import (
	"context"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v5"
)

func updateMeasurementStatistics(pg *postgres) {

	updateLowestAircraft(pg)
	updateHighestAircraft(pg)
	updateSlowestAircraft(pg)

	updateFastestAircraft(pg)

	// updateAltitudeStatistics(pg, "lowest_aircraft", "barometric_altitude", "DESC", 99999, "lowest_aircraft_processed")
	// updateAltitudeStatistics(pg, "highest_aircraft", "barometric_altitude", "ASC", 0, "highest_aircraft_processed")
	// updateSpeedStatistics(pg, "slowest_aircraft", "ground_speed", "DESC", 99999, "slowest_aircraft_processed")
	// updateSpeedStatistics(pg, "fastest_aircraft", "ground_speed", "ASC", 0, "fastest_aircraft_processed")

}

func updateLowestAircraft(pg *postgres) {

	processedMetricName := "lowest_aircraft_processed"
	tableName := "lowest_aircraft"
	metricName := "barometric_altitude"

	aircrafts := getAircraftsForMeasurementStatistics(pg, processedMetricName)

	fmt.Println("Aircrafts that have not been processed: ", len(aircrafts))
	if len(aircrafts) == 0 {
		return
	}

	lowestAircraftCeiling := getLowestAircraftCeiling(pg)

	sort.Slice(aircrafts, func(i, j int) bool {
		return aircrafts[i].AltBaro < aircrafts[j].AltBaro
	})

	var aircraftsToInsert []Aircraft

	for _, aircraft := range aircrafts {
		if aircraft.AltBaro < lowestAircraftCeiling {
			aircraftsToInsert = append(aircraftsToInsert, aircraft)
		} else {
			break
		}
	}

	batch := &pgx.Batch{}

	for _, aircraft := range aircraftsToInsert {
		insertStatement := `
			INSERT INTO lowest_aircraft (
				hex,
				flight,
				registration,
				type,
				first_seen,
				last_seen,
				barometric_altitude,
				geometric_altitude)
			VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (hex, first_seen)
			DO UPDATE SET
				barometric_altitude = EXCLUDED.barometric_altitude,
				geometric_altitude = EXCLUDED.geometric_altitude,
				last_seen = EXCLUDED.last_seen`

		batch.Queue(
			insertStatement,
			aircraft.Hex,
			aircraft.Flight,
			aircraft.R,
			aircraft.T,
			aircraft.FirstSeen,
			aircraft.LastSeen,
			aircraft.AltBaro,
			aircraft.AltGeom)
	}

	br := pg.db.SendBatch(context.Background(), batch)
	defer br.Close()

	for i := 0; i < len(aircraftsToInsert); i++ {
		_, err := br.Exec()
		if err != nil {
			fmt.Println("Unable to insert data: ", err)
		}
	}
	DeleteExcessRows(pg, tableName, metricName, "DESC", 50)

	if len(aircrafts) > 0 {
		MarkProcessed(pg, processedMetricName, aircrafts)
	}

}

func updateHighestAircraft(pg *postgres) {

	processedMetricName := "highest_aircraft_processed"
	tableName := "highest_aircraft"
	metricName := "barometric_altitude"

	aircrafts := getAircraftsForMeasurementStatistics(pg, processedMetricName)

	fmt.Println("Aircrafts that have not been processed: ", len(aircrafts))
	if len(aircrafts) == 0 {
		return
	}

	highestAircraftFloor := getHighestAircraftFloor(pg)

	sort.Slice(aircrafts, func(i, j int) bool {
		return aircrafts[i].AltBaro > aircrafts[j].AltBaro
	})

	var aircraftsToInsert []Aircraft

	for _, aircraft := range aircrafts {
		if aircraft.AltBaro > highestAircraftFloor {
			aircraftsToInsert = append(aircraftsToInsert, aircraft)
		} else {
			break
		}
	}

	batch := &pgx.Batch{}

	for _, aircraft := range aircraftsToInsert {
		insertStatement := `
			INSERT INTO highest_aircraft (
				hex,
				flight,
				registration,
				type,
				first_seen,
				last_seen,
				barometric_altitude,
				geometric_altitude)
			VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (hex, first_seen)
			DO UPDATE SET
				barometric_altitude = EXCLUDED.barometric_altitude,
				geometric_altitude = EXCLUDED.geometric_altitude,
				last_seen = EXCLUDED.last_seen`

		batch.Queue(
			insertStatement,
			aircraft.Hex,
			aircraft.Flight,
			aircraft.R,
			aircraft.T,
			aircraft.FirstSeen,
			aircraft.LastSeen,
			aircraft.AltBaro,
			aircraft.AltGeom)
	}

	br := pg.db.SendBatch(context.Background(), batch)
	defer br.Close()

	for i := 0; i < len(aircraftsToInsert); i++ {
		_, err := br.Exec()
		if err != nil {
			fmt.Println("Unable to insert data: ", err)
		}
	}

	DeleteExcessRows(pg, tableName, metricName, "ASC", 50)

	if len(aircrafts) > 0 {
		MarkProcessed(pg, processedMetricName, aircrafts)
	}
}

func updateSlowestAircraft(pg *postgres) {

	processedMetricName := "slowest_aircraft_processed"
	tableName := "slowest_aircraft"
	metricName := "ground_speed"

	aircrafts := getAircraftsForMeasurementStatistics(pg, processedMetricName)

	fmt.Println("Aircrafts that have not been processed: ", len(aircrafts))
	if len(aircrafts) == 0 {
		return
	}

	slowestAircraftCeiling := getSlowestAircraftCeiling(pg)

	sort.Slice(aircrafts, func(i, j int) bool {
		return aircrafts[i].Gs < aircrafts[j].Gs
	})

	var aircraftsToInsert []Aircraft

	for _, aircraft := range aircrafts {
		if aircraft.Gs < slowestAircraftCeiling {
			aircraftsToInsert = append(aircraftsToInsert, aircraft)
		} else {
			break
		}
	}

	batch := &pgx.Batch{}

	for _, aircraft := range aircraftsToInsert {
		insertStatement := `
			INSERT INTO slowest_aircraft (
				hex,
				flight,
				registration,
				type,
				first_seen,
				last_seen,
				ground_speed,
				indicated_air_speed,
				true_air_speed)
			VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (hex, first_seen)
			DO UPDATE SET
				ground_speed = EXCLUDED.ground_speed,
				indicated_air_speed = EXCLUDED.indicated_air_speed,
				true_air_speed = EXCLUDED.true_air_speed,
				last_seen = EXCLUDED.last_seen`

		batch.Queue(
			insertStatement,
			aircraft.Hex,
			aircraft.Flight,
			aircraft.R,
			aircraft.T,
			aircraft.FirstSeen,
			aircraft.LastSeen,
			aircraft.Gs,
			aircraft.Tas,
			aircraft.Ias)
	}

	br := pg.db.SendBatch(context.Background(), batch)
	defer br.Close()

	for i := 0; i < len(aircraftsToInsert); i++ {
		_, err := br.Exec()
		if err != nil {
			fmt.Println("Unable to insert data: ", err)
		}
	}

	DeleteExcessRows(pg, tableName, metricName, "DESC", 50)

	if len(aircrafts) > 0 {
		fmt.Println("Marking processed for slowest_aircraft")
		MarkProcessed(pg, processedMetricName, aircrafts)
	}
}

func updateFastestAircraft(pg *postgres) {

	processedMetricName := "fastest_aircraft_processed"
	// tableName := "fastest_aircraft"
	metricName := "ground_speed"

	aircrafts := getAircraftsForMeasurementStatistics(pg, processedMetricName)

	fmt.Println("Aircrafts that have not been processed: ", len(aircrafts))
	if len(aircrafts) == 0 {
		return
	}

	fastestAircraftFloor := getFastestAircraftFloor(pg)

	sort.Slice(aircrafts, func(i, j int) bool {
		return aircrafts[i].Gs > aircrafts[j].Gs
	})

	var aircraftsToInsert []Aircraft

	for _, aircraft := range aircrafts {
		if aircraft.Gs > fastestAircraftFloor {
			aircraftsToInsert = append(aircraftsToInsert, aircraft)
		} else {
			break
		}
	}

	batch := &pgx.Batch{}

	for _, aircraft := range aircraftsToInsert {
		insertStatement := `
			INSERT INTO fastest_aircraft (
				hex,
				flight,
				registration,
				type,
				first_seen,
				last_seen,
				ground_speed,
				indicated_air_speed,
				true_air_speed)
			VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (hex, first_seen)
			DO UPDATE SET
				ground_speed = EXCLUDED.ground_speed,
				indicated_air_speed = EXCLUDED.indicated_air_speed,
				true_air_speed = EXCLUDED.true_air_speed,
				last_seen = EXCLUDED.last_seen`

		batch.Queue(
			insertStatement,
			aircraft.Hex,
			aircraft.Flight,
			aircraft.R,
			aircraft.T,
			aircraft.FirstSeen,
			aircraft.LastSeen,
			aircraft.Gs,
			aircraft.Tas,
			aircraft.Ias)
	}

	br := pg.db.SendBatch(context.Background(), batch)
	defer br.Close()

	for i := 0; i < len(aircraftsToInsert); i++ {
		_, err := br.Exec()
		if err != nil {
			fmt.Println("Unable to insert data: ", err)
		}
	}

	DeleteExcessRows(pg, "fastest_aircraft", metricName, "ASC", 50)

	if len(aircrafts) > 0 {
		MarkProcessed(pg, processedMetricName, aircrafts)
	}
}

func getAircraftsForMeasurementStatistics(pg *postgres, processedColumnName string) []Aircraft {

	query := `SELECT id, hex, flight, r, t, first_seen, last_seen, alt_baro, alt_geom, gs, ias, tas
				FROM aircraft_data
				WHERE ` + processedColumnName + ` = false`

	rows, err := pg.db.Query(context.Background(), query)
	if err != nil {
		fmt.Println("Error querying db in updateMeasurementStatistics(): ", err)
	}
	defer rows.Close()

	var aircrafts []Aircraft

	for rows.Next() {

		var aircraft Aircraft

		err := rows.Scan(
			&aircraft.Id,
			&aircraft.Hex,
			&aircraft.Flight,
			&aircraft.R,
			&aircraft.T,
			&aircraft.FirstSeen,
			&aircraft.LastSeen,
			&aircraft.AltBaro,
			&aircraft.AltGeom,
			&aircraft.Gs,
			&aircraft.Ias,
			&aircraft.Tas)

		if err != nil {
			fmt.Println("Error scanning rows in updateMeasurementStatistics(): ", err)
		}
		aircrafts = append(aircrafts, aircraft)
	}

	return aircrafts
}
