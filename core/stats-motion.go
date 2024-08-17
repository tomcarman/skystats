package main

import (
	"context"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v5"
)

func updateMeasurementStatistics(pg *postgres) {

	updateAltitudeStatistics(pg, "lowest_aircraft", "barometric_altitude", "DESC", 99999, "lowest_aircraft_processed")
	updateAltitudeStatistics(pg, "highest_aircraft", "barometric_altitude", "ASC", 0, "highest_aircraft_processed")

	updateSpeedStatistics(pg, "slowest_aircraft", "ground_speed", "DESC", 99999, "slowest_aircraft_processed")
	updateSpeedStatistics(pg, "fastest_aircraft", "ground_speed", "ASC", 0, "fastest_aircraft_processed")

}

func updateSpeedStatistics(pg *postgres, tableName string, metricName string, sortOrder string, defaultValue float64, processedColumnName string) {

	aircrafts := getAircraftsForMeasurementStatistics(pg, processedColumnName)

	// Get the fastest or slowest based on ground speed
	fastestOrSlowest, ok := GetLowestOrHighest(pg, tableName, "ground_speed", sortOrder).(float64)
	if !ok {
		fastestOrSlowest = float64(defaultValue)
	}

	// Sort aircrafts based on the sortOrder
	if sortOrder == "ASC" {
		sort.Slice(aircrafts, func(i, j int) bool {
			return aircrafts[i].Gs > aircrafts[j].Gs
		})
	} else {
		sort.Slice(aircrafts, func(i, j int) bool {
			return aircrafts[i].Gs < aircrafts[j].Gs
		})
	}

	var aircraftsToInsert []Aircraft

	for _, aircraft := range aircrafts {
		if (sortOrder == "ASC" && aircraft.Gs > fastestOrSlowest) || (sortOrder == "DESC" && aircraft.Gs < fastestOrSlowest) {
			aircraftsToInsert = append(aircraftsToInsert, aircraft)
		} else {
			break
		}
	}

	batch := &pgx.Batch{}

	for _, aircraft := range aircraftsToInsert {
		insertStatement := fmt.Sprintf(`
								INSERT INTO %s (
									hex, flight, registration, type, first_seen, last_seen,
									ground_speed, indicated_air_speed, true_air_speed)
								VALUES (
									$1, $2, $3, $4, $5, $6, $7, $8, $9)
								ON CONFLICT (hex, first_seen)
								DO UPDATE SET
									ground_speed = EXCLUDED.ground_speed,
									indicated_air_speed = EXCLUDED.indicated_air_speed,
									true_air_speed = EXCLUDED.true_air_speed,
									last_seen = EXCLUDED.last_seen`, tableName)

		batch.Queue(
			insertStatement,
			aircraft.Hex, aircraft.Flight, aircraft.R, aircraft.T, aircraft.FirstSeen, aircraft.LastSeen,
			aircraft.Gs, aircraft.Tas, aircraft.Ias)
	}

	br := pg.db.SendBatch(context.Background(), batch)
	defer br.Close()

	for i := 0; i < len(aircraftsToInsert); i++ {
		_, err := br.Exec()
		if err != nil {
			fmt.Println("Unable to insert data: ", err)
		}
	}

	DeleteExcessRows(pg, tableName, metricName, sortOrder, 50)

	if len(aircrafts) > 0 {
		MarkProcessed(pg, processedColumnName, aircrafts)
	}
}

func updateAltitudeStatistics(pg *postgres, tableName string, metricName string, sortOrder string, defaultValue int, processedColumnName string) {

	aircrafts := getAircraftsForMeasurementStatistics(pg, processedColumnName)

	// Get the highest or lowest barometric altitude
	highestOrLowestAltBaro, ok := GetLowestOrHighest(pg, tableName, "barometric_altitude", sortOrder).(int32)
	if !ok {
		highestOrLowestAltBaro = int32(defaultValue)
	}

	// Sort aircrafts based on the sortOrder
	if sortOrder == "ASC" {
		sort.Slice(aircrafts, func(i, j int) bool {
			return aircrafts[i].AltBaro > aircrafts[j].AltBaro
		})
	} else {
		sort.Slice(aircrafts, func(i, j int) bool {
			return aircrafts[i].AltBaro < aircrafts[j].AltBaro
		})
	}

	var aircraftsToInsert []Aircraft

	for _, aircraft := range aircrafts {
		if (sortOrder == "ASC" && aircraft.AltBaro > int(highestOrLowestAltBaro)) || (sortOrder == "DESC" && aircraft.AltBaro < int(highestOrLowestAltBaro)) {
			aircraftsToInsert = append(aircraftsToInsert, aircraft)
		} else {
			break
		}
	}

	batch := &pgx.Batch{}

	for _, aircraft := range aircraftsToInsert {
		insertStatement := fmt.Sprintf(`
								INSERT INTO %s (
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
									last_seen = EXCLUDED.last_seen`, tableName)

		batch.Queue(
			insertStatement,
			aircraft.Hex, aircraft.Flight, aircraft.R, aircraft.T, aircraft.FirstSeen, aircraft.LastSeen,
			aircraft.AltBaro, aircraft.AltGeom)
	}

	br := pg.db.SendBatch(context.Background(), batch)
	defer br.Close()

	for i := 0; i < len(aircraftsToInsert); i++ {
		_, err := br.Exec()
		if err != nil {
			fmt.Println("Unable to insert data: ", err)
		}
	}

	DeleteExcessRows(pg, tableName, metricName, sortOrder, 50)

	if len(aircrafts) > 0 {
		MarkProcessed(pg, processedColumnName, aircrafts)
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
