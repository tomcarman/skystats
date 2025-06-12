package main

import (
	"context"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v5"
)

type MetricType string

const (
	MetricLowest  MetricType = "lowest"
	MetricHighest MetricType = "highest"
	MetricFastest MetricType = "fastest"
	MetricSlowest MetricType = "slowest"
)

type StatisticConfig struct {
	MetricType          MetricType
	ProcessedFieldName  string
	TableName          string
	MetricColumnName   string
	SortAscending      bool
	DefaultThreshold   interface{}
	GetThresholdFunc   func(pg *postgres) interface{}
	GetMetricValue     func(aircraft Aircraft) interface{}
	ValidateMetric     func(value interface{}) bool
	BuildInsertSQL     func() string
	GetInsertParams    func(aircraft Aircraft) []interface{}
}

func updateMeasurementStatistics(pg *postgres) {
	aircrafts := getAircraftsForMeasurementStatistics(pg)

	configs := getStatisticConfigs()
	for _, config := range configs {
		updateStatistic(pg, aircrafts, config)
	}
}

func updateStatistic(pg *postgres, aircrafts []Aircraft, config StatisticConfig) {
	var aircraftToProcess []Aircraft

	for _, aircraft := range aircrafts {
		if !getProcessedFlag(aircraft, config.MetricType) {
			aircraftToProcess = append(aircraftToProcess, aircraft)
		}
	}

	if len(aircraftToProcess) == 0 {
		return
	}

	threshold := config.GetThresholdFunc(pg)

	sort.Slice(aircraftToProcess, func(i, j int) bool {
		val1 := config.GetMetricValue(aircraftToProcess[i])
		val2 := config.GetMetricValue(aircraftToProcess[j])
		
		if config.SortAscending {
			return compareValues(val1, val2, true)
		}
		return compareValues(val1, val2, false)
	})

	var aircraftsToInsert []Aircraft
	for _, aircraft := range aircraftToProcess {
		metricValue := config.GetMetricValue(aircraft)
		
		if !config.ValidateMetric(metricValue) {
			continue
		}
		
		if shouldInsert(metricValue, threshold, config.SortAscending) {
			aircraftsToInsert = append(aircraftsToInsert, aircraft)
		} else {
			break
		}
	}

	if len(aircraftsToInsert) > 0 {
		insertStatistics(pg, aircraftsToInsert, config)
		DeleteExcessRows(pg, config.TableName, config.MetricColumnName, getSortOrder(config.SortAscending), 50)
	}

	if len(aircraftToProcess) > 0 {
		MarkProcessed(pg, config.ProcessedFieldName, aircraftToProcess)
	}
}

func getProcessedFlag(aircraft Aircraft, metricType MetricType) bool {
	switch metricType {
	case MetricLowest:
		return aircraft.LowestProcessed
	case MetricHighest:
		return aircraft.HighestProcessed
	case MetricFastest:
		return aircraft.FastestProcessed
	case MetricSlowest:
		return aircraft.SlowestProcessed
	default:
		return true
	}
}

func compareValues(val1, val2 interface{}, ascending bool) bool {
	switch v1 := val1.(type) {
	case int:
		v2 := val2.(int)
		if ascending {
			return v1 < v2
		}
		return v1 > v2
	case float64:
		v2 := val2.(float64)
		if ascending {
			return v1 < v2
		}
		return v1 > v2
	default:
		return false
	}
}

func shouldInsert(metricValue, threshold interface{}, ascending bool) bool {
	switch mv := metricValue.(type) {
	case int:
		th := threshold.(int)
		if ascending {
			return mv < th
		}
		return mv > th
	case float64:
		th := threshold.(float64)
		if ascending {
			return mv < th
		}
		return mv > th
	default:
		return false
	}
}

func getSortOrder(ascending bool) string {
	if ascending {
		return "ASC"
	}
	return "DESC"
}

func insertStatistics(pg *postgres, aircrafts []Aircraft, config StatisticConfig) {
	batch := &pgx.Batch{}

	for _, aircraft := range aircrafts {
		batch.Queue(config.BuildInsertSQL(), config.GetInsertParams(aircraft)...)
	}

	br := pg.db.SendBatch(context.Background(), batch)
	defer br.Close()

	for i := 0; i < len(aircrafts); i++ {
		_, err := br.Exec()
		if err != nil {
			fmt.Printf("insertStatistics() - Unable to insert %s data: %v\n", config.MetricType, err)
		}
	}
}

func getStatisticConfigs() []StatisticConfig {
	return []StatisticConfig{
		{
			MetricType:         MetricLowest,
			ProcessedFieldName: "lowest_aircraft_processed",
			TableName:         "lowest_aircraft",
			MetricColumnName:  "barometric_altitude",
			SortAscending:     true,
			DefaultThreshold:  999999,
			GetThresholdFunc: func(pg *postgres) interface{} {
				return getLowestAircraftCeiling(pg)
			},
			GetMetricValue: func(aircraft Aircraft) interface{} {
				return aircraft.AltBaro
			},
			ValidateMetric: func(value interface{}) bool {
				return value.(int) >= 1
			},
			BuildInsertSQL: func() string {
				return `INSERT INTO lowest_aircraft (
					hex, flight, registration, type, first_seen, last_seen, 
					barometric_altitude, geometric_altitude)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
				ON CONFLICT (hex, first_seen)
				DO UPDATE SET
					barometric_altitude = EXCLUDED.barometric_altitude,
					geometric_altitude = EXCLUDED.geometric_altitude,
					last_seen = EXCLUDED.last_seen`
			},
			GetInsertParams: func(aircraft Aircraft) []interface{} {
				return []interface{}{
					aircraft.Hex, aircraft.Flight, aircraft.R, aircraft.T,
					aircraft.FirstSeen, aircraft.LastSeen, aircraft.AltBaro, aircraft.AltGeom,
				}
			},
		},
		{
			MetricType:         MetricHighest,
			ProcessedFieldName: "highest_aircraft_processed",
			TableName:         "highest_aircraft",
			MetricColumnName:  "barometric_altitude",
			SortAscending:     false,
			DefaultThreshold:  0,
			GetThresholdFunc: func(pg *postgres) interface{} {
				return getHighestAircraftFloor(pg)
			},
			GetMetricValue: func(aircraft Aircraft) interface{} {
				return aircraft.AltBaro
			},
			ValidateMetric: func(value interface{}) bool {
				return value.(int) >= 1
			},
			BuildInsertSQL: func() string {
				return `INSERT INTO highest_aircraft (
					hex, flight, registration, type, first_seen, last_seen, 
					barometric_altitude, geometric_altitude)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
				ON CONFLICT (hex, first_seen)
				DO UPDATE SET
					barometric_altitude = EXCLUDED.barometric_altitude,
					geometric_altitude = EXCLUDED.geometric_altitude,
					last_seen = EXCLUDED.last_seen`
			},
			GetInsertParams: func(aircraft Aircraft) []interface{} {
				return []interface{}{
					aircraft.Hex, aircraft.Flight, aircraft.R, aircraft.T,
					aircraft.FirstSeen, aircraft.LastSeen, aircraft.AltBaro, aircraft.AltGeom,
				}
			},
		},
		{
			MetricType:         MetricFastest,
			ProcessedFieldName: "fastest_aircraft_processed",
			TableName:         "fastest_aircraft",
			MetricColumnName:  "ground_speed",
			SortAscending:     false,
			DefaultThreshold:  0.0,
			GetThresholdFunc: func(pg *postgres) interface{} {
				return getFastestAircraftFloor(pg)
			},
			GetMetricValue: func(aircraft Aircraft) interface{} {
				return aircraft.Gs
			},
			ValidateMetric: func(value interface{}) bool {
				return value.(float64) >= 1.0
			},
			BuildInsertSQL: func() string {
				return `INSERT INTO fastest_aircraft (
					hex, flight, registration, type, first_seen, last_seen, 
					ground_speed, indicated_air_speed, true_air_speed)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
				ON CONFLICT (hex, first_seen)
				DO UPDATE SET
					ground_speed = EXCLUDED.ground_speed,
					indicated_air_speed = EXCLUDED.indicated_air_speed,
					true_air_speed = EXCLUDED.true_air_speed,
					last_seen = EXCLUDED.last_seen`
			},
			GetInsertParams: func(aircraft Aircraft) []interface{} {
				return []interface{}{
					aircraft.Hex, aircraft.Flight, aircraft.R, aircraft.T,
					aircraft.FirstSeen, aircraft.LastSeen, aircraft.Gs, aircraft.Ias, aircraft.Tas,
				}
			},
		},
		{
			MetricType:         MetricSlowest,
			ProcessedFieldName: "slowest_aircraft_processed",
			TableName:         "slowest_aircraft",
			MetricColumnName:  "ground_speed",
			SortAscending:     true,
			DefaultThreshold:  99999.0,
			GetThresholdFunc: func(pg *postgres) interface{} {
				return getSlowestAircraftCeiling(pg)
			},
			GetMetricValue: func(aircraft Aircraft) interface{} {
				return aircraft.Gs
			},
			ValidateMetric: func(value interface{}) bool {
				return value.(float64) >= 1.0
			},
			BuildInsertSQL: func() string {
				return `INSERT INTO slowest_aircraft (
					hex, flight, registration, type, first_seen, last_seen, 
					ground_speed, indicated_air_speed, true_air_speed)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
				ON CONFLICT (hex, first_seen)
				DO UPDATE SET
					ground_speed = EXCLUDED.ground_speed,
					indicated_air_speed = EXCLUDED.indicated_air_speed,
					true_air_speed = EXCLUDED.true_air_speed,
					last_seen = EXCLUDED.last_seen`
			},
			GetInsertParams: func(aircraft Aircraft) []interface{} {
				return []interface{}{
					aircraft.Hex, aircraft.Flight, aircraft.R, aircraft.T,
					aircraft.FirstSeen, aircraft.LastSeen, aircraft.Gs, aircraft.Ias, aircraft.Tas,
				}
			},
		},
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

	fmt.Println("Aircrafts that have not have statistics processed: ", len(aircrafts))
	return aircrafts
}
