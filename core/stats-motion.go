package main

import (
	"context"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v5"
)

func updateMeasurementStatistics(pg *postgres) {

	aircrafts := getAircraftsForMeasurementStatistics(pg)

	updateLowestAircraft(pg, aircrafts)
	updateFastestAircraft(pg, aircrafts)
	updateHighestAircraft(pg, aircrafts)
	updateSlowestAircraft(pg, aircrafts)

}

func updateLowestAircraft(pg *postgres, aircrafts []Aircraft) {
	processedMetricName := "lowest_aircraft_processed"
	tableName := "lowest_aircraft"
	metricName := "barometric_altitude"

	var aircraftToProcess []Aircraft

	for _, aircraft := range aircrafts {
		if !aircraft.LowestProcessed {
			aircraftToProcess = append(aircraftToProcess, aircraft)
		}
	}

	if len(aircraftToProcess) == 0 {
		return
	}

	// fmt.Println("updateLowestAircraft() - Aircraft to process: ", len(aircraftToProcess))

	lowestAircraftCeiling := getLowestAircraftCeiling(pg)

	sort.Slice(aircraftToProcess, func(i, j int) bool {
		return aircraftToProcess[i].AltBaro < aircraftToProcess[j].AltBaro
	})

	var aircraftsToInsert []Aircraft

	for _, aircraft := range aircraftToProcess {
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
			fmt.Println("updateLowestAircraft() - Unable to insert data: ", err)
		}
	}
	DeleteExcessRows(pg, tableName, metricName, "DESC", 50)

	if len(aircraftToProcess) > 0 {
		MarkProcessed(pg, processedMetricName, aircraftToProcess)
	}

}

func updateHighestAircraft(pg *postgres, aircrafts []Aircraft) {

	processedMetricName := "highest_aircraft_processed"
	tableName := "highest_aircraft"
	metricName := "barometric_altitude"

	var aircraftToProcess []Aircraft

	for _, aircraft := range aircrafts {
		if !aircraft.HighestProcessed {
			aircraftToProcess = append(aircraftToProcess, aircraft)
		}
	}

	if len(aircraftToProcess) == 0 {
		return
	}

	// fmt.Println("updateHighestAircraft() - Aircraft to process: ", len(aircraftToProcess))

	highestAircraftFloor := getHighestAircraftFloor(pg)

	sort.Slice(aircraftToProcess, func(i, j int) bool {
		return aircraftToProcess[i].AltBaro > aircraftToProcess[j].AltBaro
	})

	var aircraftsToInsert []Aircraft

	for _, aircraft := range aircraftToProcess {
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
			fmt.Println("updateHighestAircraft() - Unable to insert data: ", err)
		}
	}

	DeleteExcessRows(pg, tableName, metricName, "ASC", 50)

	if len(aircraftToProcess) > 0 {
		MarkProcessed(pg, processedMetricName, aircraftToProcess)
	}
}

func updateSlowestAircraft(pg *postgres, aircrafts []Aircraft) {

	processedMetricName := "slowest_aircraft_processed"
	tableName := "slowest_aircraft"
	metricName := "ground_speed"

	var aircraftToProcess []Aircraft

	for _, aircraft := range aircrafts {
		if !aircraft.SlowestProcessed {
			aircraftToProcess = append(aircraftToProcess, aircraft)
		}
	}

	if len(aircraftToProcess) == 0 {
		return
	}

	// fmt.Println("updateSlowestAircraft() - Aircraft to process: ", len(aircraftToProcess))

	slowestAircraftCeiling := getSlowestAircraftCeiling(pg)

	sort.Slice(aircraftToProcess, func(i, j int) bool {
		return aircraftToProcess[i].Gs < aircraftToProcess[j].Gs
	})

	var aircraftsToInsert []Aircraft

	for _, aircraft := range aircraftToProcess {
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
			fmt.Println("updateSlowestAircraft() - Unable to insert data: ", err)
		}
	}

	DeleteExcessRows(pg, tableName, metricName, "DESC", 50)

	if len(aircraftToProcess) > 0 {
		MarkProcessed(pg, processedMetricName, aircraftToProcess)
	}
}

func updateFastestAircraft(pg *postgres, aircrafts []Aircraft) {

	processedMetricName := "fastest_aircraft_processed"
	tableName := "fastest_aircraft"
	metricName := "ground_speed"

	var aircraftToProcess []Aircraft

	for _, aircraft := range aircrafts {
		if !aircraft.FastestProcessed {
			aircraftToProcess = append(aircraftToProcess, aircraft)
		}
	}

	if len(aircraftToProcess) == 0 {
		return
	}

	// fmt.Println("updateFastestAircraft() - Aircraft to process: ", len(aircraftToProcess))

	fastestAircraftFloor := getFastestAircraftFloor(pg)

	sort.Slice(aircraftToProcess, func(i, j int) bool {
		return aircraftToProcess[i].Gs > aircraftToProcess[j].Gs
	})

	var aircraftsToInsert []Aircraft

	for _, aircraft := range aircraftToProcess {
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
			fmt.Println("updateFastestAircraft() - Unable to insert data: ", err)
		}
	}

	DeleteExcessRows(pg, tableName, metricName, "ASC", 50)

	if len(aircraftToProcess) > 0 {
		MarkProcessed(pg, processedMetricName, aircraftToProcess)
	}
}

func getAircraftsForMeasurementStatistics(pg *postgres) []Aircraft {

	query := `SELECT id, hex, flight, r, t, first_seen, last_seen, alt_baro, alt_geom, gs, ias, tas, 
				lowest_aircraft_processed, highest_aircraft_processed, fastest_aircraft_processed, slowest_aircraft_processed
				FROM aircraft_data
				WHERE lowest_aircraft_processed = false OR
					highest_aircraft_processed = false OR
					fastest_aircraft_processed = false OR
					slowest_aircraft_processed = false`

	rows, err := pg.db.Query(context.Background(), query)
	if err != nil {
		fmt.Println("getAircraftsForMeasurementStatistics() - Error querying db: ", err)
		return nil
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
			&aircraft.Tas,
			&aircraft.LowestProcessed,
			&aircraft.HighestProcessed,
			&aircraft.FastestProcessed,
			&aircraft.SlowestProcessed)

		if err != nil {
			fmt.Println("getAircraftsForMeasurementStatistics() - Error scanning rows: ", err)
			return nil
		}
		aircrafts = append(aircrafts, aircraft)
	}

	fmt.Println("Aircrafts that have not been processed: ", len(aircrafts))
	return aircrafts
}
