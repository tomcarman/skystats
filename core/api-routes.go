package main

import (
	"context"
	"fmt"
)

func getAircraftWithoutRegistrationProcessed(pg *postgres) []Aircraft {

	query := `
		SELECT id, hex
		FROM aircraft_data
		WHERE 
			hex != '' AND
			registration_processed = false
		ORDER BY first_seen ASC
		LIMIT 10`

	rows, err := pg.db.Query(context.Background(), query)

	if err != nil {
		fmt.Println("getAircraftWithoutRegistrationProcessed() - Error querying db: ", err)
		return nil
	}
	defer rows.Close()

	var aircrafts []Aircraft

	for rows.Next() {

		var aircraft Aircraft

		err := rows.Scan(
			&aircraft.Id,
			&aircraft.Hex,
		)

		if err != nil {
			fmt.Println("getAircraftWithoutRegistrationProcessed() - Error scanning rows: ", err)
			return nil
		}

		aircrafts = append(aircrafts, aircraft)
	}

	fmt.Println("Aircrafts that have not have route processed: ", len(aircrafts))
	return aircrafts
}

/* TODO
- check if hex already in the aircraft_registration table
	- if so, update registration_processed to true
	- if not, create list to send to adsbdb.com

*/
