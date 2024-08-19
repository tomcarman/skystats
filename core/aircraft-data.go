package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	cheapruler "github.com/JamesLMilner/cheap-ruler-go"
	"github.com/jackc/pgx/v5"
)

func updateAircraftDatabase(pg *postgres) {

	responseData := Fetch()

	var response Response
	json.Unmarshal(responseData, &response)

	loc := []float64{getLat(), getLon()}

	var aircraftsInRange []Aircraft

	for _, aircraft := range response.Aircraft {

		planeLoc := []float64{aircraft.Lat, aircraft.Lon}
		distance := getRuler().Distance(loc, planeLoc)

		if distance < getRadius() {
			aircraftsInRange = append(aircraftsInRange, aircraft)
		}
	}
	pg.updateDatabase(response.Now, aircraftsInRange)
}

func getRuler() *cheapruler.CheapRuler {
	ruler, err := cheapruler.NewCheapruler(getLon(), "kilometers")
	if err != nil {
		fmt.Println("Error creating ruler: ", err)
	}

	return &ruler
}

func (pg *postgres) updateDatabase(nowEpoch float64, aircrafts []Aircraft) {

	existingAircrafts := getAircraftsRecentlySeen(pg, nowEpoch, aircrafts)

	if len(existingAircrafts) > 0 {
		updateExistingAircrafts(pg, nowEpoch, aircrafts, existingAircrafts)
	}

	insertNewAircrafts(pg, nowEpoch, existingAircrafts, aircrafts)

}

func getAircraftsRecentlySeen(pg *postgres, nowEpoch float64, aircrafts []Aircraft) map[string]*Aircraft {

	existingAircrafts := make(map[string]*Aircraft)

	var hexValues []string
	for _, a := range aircrafts {
		hexValues = append(hexValues, a.Hex)
	}

	query := `
		SELECT DISTINCT ON (hex)
			id,
			hex,
			last_seen_epoch,
			alt_baro,
			alt_geom,
			gs,
			ias,
			tas
		FROM aircraft_data
		WHERE hex = ANY($1::text[])
		ORDER BY hex, last_seen DESC;
    `

	rows, err := pg.db.Query(context.Background(), query, hexValues)
	if err != nil {
		fmt.Println("getAircraftsRecentlySeen() - Error querying db: ", err)
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		var existingAircraft Aircraft
		err := rows.Scan(
			&existingAircraft.Id,
			&existingAircraft.Hex,
			&existingAircraft.LastSeenEpoch,
			&existingAircraft.AltBaro,
			&existingAircraft.AltGeom,
			&existingAircraft.Gs,
			&existingAircraft.Ias, &existingAircraft.Tas)

		if err != nil {
			fmt.Println("getAircraftsRecentlySeen() - Error scanning rows: ", err)
			continue
		}
		if nowEpoch-existingAircraft.LastSeenEpoch > 300 {
			continue
		}

		existingAircrafts[existingAircraft.Hex] = &existingAircraft
	}

	return existingAircrafts

}

func insertNewAircrafts(pg *postgres, nowEpoch float64, existingAircrafts map[string]*Aircraft, aircrafts []Aircraft) {

	batch := &pgx.Batch{}

	nowAsTime := time.Unix(int64(nowEpoch), 0)
	nowAsEpoch := int64(nowEpoch)

	var aircraftsToInsert []Aircraft

	for _, aircraft := range aircrafts {
		_, exists := existingAircrafts[aircraft.Hex]
		if !exists {
			aircraftsToInsert = append(aircraftsToInsert, aircraft)
			insertStatement := `
				INSERT INTO aircraft_data (
					hex, flight,
					first_seen,
					first_seen_epoch,
					last_seen,
					last_seen_epoch,
					type,
					r,
					t,
					alt_baro,
					alt_geom,
					gs,
					ias,
					tas,
					track,
					baro_rate,
					nav_qnh,
					nav_altitude_mcp,
					nav_heading,
					lat,
					lon,
					nic,
					rc,
					seen_pos,
					r_dst,
					r_dir,
					version,
					nic_baro,
					nac_p,
					nac_v,
					sil,
					sil_type,
					alert,
					spi,
					mlat,
					tisb,
					messages,
					seen,
					rssi
				) VALUES (
					$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, 
					$16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, 
					$29, $30, $31, $32, $33, $34, $35, $36, $37, $38, $39
				)`

			batch.Queue(insertStatement,
				aircraft.Hex,
				aircraft.Flight,
				nowAsTime,
				nowAsEpoch,
				nowAsTime,
				nowAsEpoch,
				aircraft.Type,
				aircraft.R,
				aircraft.T,
				aircraft.AltBaro,
				aircraft.AltGeom,
				aircraft.Gs,
				aircraft.Ias,
				aircraft.Tas,
				aircraft.Track,
				aircraft.BaroRate,
				aircraft.NavQnh,
				aircraft.NavAltitudeMcp,
				aircraft.NavHeading,
				aircraft.Lat,
				aircraft.Lon,
				aircraft.Nic,
				aircraft.Rc,
				aircraft.SeenPos,
				aircraft.RDst,
				aircraft.RDir,
				aircraft.Version,
				aircraft.NicBaro,
				aircraft.NacP,
				aircraft.NacV,
				aircraft.Sil,
				aircraft.SilType,
				aircraft.Alert,
				aircraft.Spi,
				aircraft.Mlat,
				aircraft.Tisb,
				aircraft.Messages,
				aircraft.Seen,
				aircraft.Rssi)
		}
	}

	br := pg.db.SendBatch(context.Background(), batch)
	defer br.Close()

	for i := 0; i < len(aircraftsToInsert); i++ {
		_, err := br.Exec()
		if err != nil {
			fmt.Println("insertNewAircrafts() - unable to insert data: ", err)
		}
	}
}

func updateExistingAircrafts(pg *postgres, nowEpoch float64, aircrafts []Aircraft, existingAircrafts map[string]*Aircraft) {

	batch := &pgx.Batch{}

	for _, aircraft := range aircrafts {
		existingAircraft, exists := existingAircrafts[aircraft.Hex]
		if !exists {
			continue
		}

		// Update last seen time
		existingAircraft.LastSeenEpoch = nowEpoch
		existingAircraft.LastSeen = time.Unix(int64(nowEpoch), 0)

		// Update barometric altitude & geometric altitudes if higher than already stored
		if existingAircraft.AltBaro < aircraft.AltBaro {
			existingAircraft.AltBaro = aircraft.AltBaro
		}
		if existingAircraft.AltGeom < aircraft.AltGeom {
			existingAircraft.AltGeom = aircraft.AltGeom
		}

		// Update ground speed, indicated air speed, and true air speed if higher than already stored
		if existingAircraft.Gs < aircraft.Gs {
			existingAircraft.Gs = aircraft.Gs
		}
		if existingAircraft.Ias < aircraft.Ias {
			existingAircraft.Ias = aircraft.Ias
		}
		if existingAircraft.Tas < aircraft.Tas {
			existingAircraft.Tas = aircraft.Tas
		}

		updateStatement := `UPDATE aircraft_data
							SET last_seen = $1,
								last_seen_epoch = $2,
								alt_baro = $3,
								alt_geom = $4,
								gs = $5,
								ias = $6,
								tas = $7
							WHERE id = $8`

		batch.Queue(
			updateStatement,
			existingAircraft.LastSeen,
			existingAircraft.LastSeenEpoch,
			existingAircraft.AltBaro,
			existingAircraft.AltGeom,
			existingAircraft.Gs,
			existingAircraft.Ias,
			existingAircraft.Tas,
			existingAircraft.Id)
	}

	br := pg.db.SendBatch(context.Background(), batch)
	defer br.Close()

	for i := 0; i < len(existingAircrafts); i++ {
		_, err := br.Exec()
		if err != nil {
			fmt.Println("updateExistingAircrafts() - unable to update data: ", err)
		}
	}
}

func getLat() float64 {
	lat, err := strconv.ParseFloat(os.Getenv("LATITUDE"), 64)
	if err != nil {
		fmt.Println("Error parsing LATITUDE: ", err)
	}
	return lat
}

func getLon() float64 {
	lon, err := strconv.ParseFloat(os.Getenv("LONGITUDE"), 64)
	if err != nil {
		fmt.Println("Error parsing LONGITUDE: ", err)
	}
	return lon
}

func getRadius() float64 {
	radius, err := strconv.ParseFloat(os.Getenv("RADIUS"), 64)
	if err != nil {
		fmt.Println("Error parsing RADIUS: ", err)
	}
	return radius
}
