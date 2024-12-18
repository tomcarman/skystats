package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

func updateInterestingSeen(pg *postgres) {

	aircrafts := unprocessedInteresting(pg)

	if len(aircrafts) == 0 {
		return
	}

	if len(aircrafts) > 1000 {
		aircrafts = aircrafts[:1000]
	}

	aircraftsMap := make(map[string]Aircraft)
	aircraftsHex := make([]string, 0, len(aircrafts))

	for _, aircraft := range aircrafts {
		aircraftsMap[strings.ToUpper(aircraft.Hex)] = aircraft
		aircraftsHex = append(aircraftsHex, aircraft.Hex)
	}

	query := `
		SELECT 
			icao,
			registration,
			operator,
			type,
			icao_type,
			"group",
			tag1,
			tag2,
			tag3,
			category,
			link,
			image_link_1,
			image_link_2,
			image_link_3,
			image_link_4
		FROM interesting_aircraft
		WHERE icao = ANY($1::text[])`

	rows, err := pg.db.Query(context.Background(), query, aircraftsHex)

	if err != nil {
		fmt.Println("updateInterestingSeen() - Error querying db: ", err)
		return
	}

	defer rows.Close()

	var interestingAircrafts []InterestingAircraft

	for rows.Next() {
		var interestingAircraft InterestingAircraft
		err := rows.Scan(
			&interestingAircraft.Icao,
			&interestingAircraft.Registration,
			&interestingAircraft.Operator,
			&interestingAircraft.Type,
			&interestingAircraft.IcaoType,
			&interestingAircraft.Group,
			&interestingAircraft.Tag1,
			&interestingAircraft.Tag2,
			&interestingAircraft.Tag3,
			&interestingAircraft.Category,
			&interestingAircraft.Link,
			&interestingAircraft.ImageLink1,
			&interestingAircraft.ImageLink2,
			&interestingAircraft.ImageLink3,
			&interestingAircraft.ImageLink4,
		)

		if err != nil {
			fmt.Println("updateInterestingSeen() - Error scanning rows: ", err)
			continue
		}

		interestingAircrafts = append(interestingAircrafts, interestingAircraft)
	}

	// Add aircraft data to interestingAircrafts
	for i := range interestingAircrafts {
		interestingAircraft := &interestingAircrafts[i]
		if aircraft, ok := aircraftsMap[interestingAircraft.Icao]; ok {
			interestingAircraft.Hex = aircraft.Hex
			interestingAircraft.Flight = aircraft.Flight
			interestingAircraft.R = aircraft.R
			interestingAircraft.T = aircraft.T
			interestingAircraft.AltBaro = aircraft.AltBaro
			interestingAircraft.AltGeom = aircraft.AltGeom
			interestingAircraft.Gs = aircraft.Gs
			interestingAircraft.Ias = aircraft.Ias
			interestingAircraft.Tas = aircraft.Tas
			interestingAircraft.Track = aircraft.Track
			interestingAircraft.BaroRate = aircraft.BaroRate
			interestingAircraft.Lat = aircraft.Lat
			interestingAircraft.Lon = aircraft.Lon
			interestingAircraft.Alert = aircraft.Alert
			interestingAircraft.DbFlags = aircraft.DbFlags
			interestingAircraft.Seen = aircraft.FirstSeen
			interestingAircraft.SeenEpoch = aircraft.FirstSeenEpoch
		}
	}

	fmt.Println("Interesting aircrafts found: ", len(interestingAircrafts))

	// Insert interesting aircraft into interesting_aircraft_seen

	batch := &pgx.Batch{}

	for _, aircraft := range interestingAircrafts {
		insertStatement := `
			INSERT INTO interesting_aircraft_seen (
				icao,
				registration,
				operator,
				type,
				icao_type,
				"group",
				tag1,
				tag2,
				tag3,
				category,
				link,
				image_link_1,
				image_link_2,
				image_link_3,
				image_link_4,
				hex,
				flight,
				r,
				t,
				alt_baro,
				alt_geom,
				gs,
				ias,
				tas,
				track,
				baro_rate,
				lat,
				lon,
				alert,
				db_flags,
				seen,
				seen_epoch)
			VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
				$11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
				$21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32)`

		batch.Queue(insertStatement,
			aircraft.Icao,
			aircraft.Registration,
			aircraft.Operator,
			aircraft.Type,
			aircraft.IcaoType,
			aircraft.Group,
			aircraft.Tag1,
			aircraft.Tag2,
			aircraft.Tag3,
			aircraft.Category,
			aircraft.Link,
			aircraft.ImageLink1,
			aircraft.ImageLink2,
			aircraft.ImageLink3,
			aircraft.ImageLink4,
			aircraft.Hex,
			aircraft.Flight,
			aircraft.R,
			aircraft.T,
			aircraft.AltBaro,
			aircraft.AltGeom,
			aircraft.Gs,
			aircraft.Ias,
			aircraft.Tas,
			aircraft.Track,
			aircraft.BaroRate,
			aircraft.Lat,
			aircraft.Lon,
			aircraft.Alert,
			aircraft.DbFlags,
			aircraft.Seen,
			aircraft.SeenEpoch)
	}

	br := pg.db.SendBatch(context.Background(), batch)
	defer br.Close()

	for i := 0; i < len(interestingAircrafts); i++ {
		_, err := br.Exec()
		if err != nil {
			fmt.Println("insertRegistrations() - Unable to insert data: ", err)
		}
	}

	// Mark all as processed
	MarkProcessed(pg, "interesting_processed", aircrafts)

}

// Mark all as processed

func unprocessedInteresting(pg *postgres) []Aircraft {

	query := `
		SELECT id,
				hex,
				flight,
				r,
				t,
				alt_baro,
				alt_geom,
				gs,
				ias,
				tas,
				track,
				baro_rate,
				lat,
				lon,
				alert,
				db_flags,
				first_seen,
				first_seen_epoch
		FROM aircraft_data
		WHERE
			hex != '' AND
			interesting_processed = false
		ORDER BY first_seen ASC`

	rows, err := pg.db.Query(context.Background(), query)

	if err != nil {
		fmt.Println("unprocessedInteresting() - Error querying db: ", err)
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
			&aircraft.AltBaro,
			&aircraft.AltGeom,
			&aircraft.Gs,
			&aircraft.Ias,
			&aircraft.Tas,
			&aircraft.Track,
			&aircraft.BaroRate,
			&aircraft.Lat,
			&aircraft.Lon,
			&aircraft.Alert,
			&aircraft.DbFlags,
			&aircraft.FirstSeen,
			&aircraft.FirstSeenEpoch,
		)

		if err != nil {
			fmt.Println("unprocessedInteresting() - Error scanning rows: ", err)
			return nil
		}

		aircrafts = append(aircrafts, aircraft)
	}

	fmt.Println("Aircrafts that have not have interesting processed: ", len(aircrafts))
	return aircrafts
}
