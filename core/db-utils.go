package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func MarkProcessed(pg *postgres, colName string, aircrafts []Aircraft) {

	fmt.Println("Entered MarkProcessed() for colName: ", colName)
	fmt.Println("aircrafts: ", len(aircrafts))

	batch := &pgx.Batch{}

	for _, aircraft := range aircrafts {
		fmt.Println("aircraft object: ", aircraft)
		updateStatement := `UPDATE aircraft_data SET ` + colName + ` = true WHERE id = $1`
		batch.Queue(updateStatement, aircraft.Id)
	}

	br := pg.db.SendBatch(context.Background(), batch)
	defer br.Close()

	for i := 0; i < len(aircrafts); i++ {
		_, err := br.Exec()
		if err != nil {
			fmt.Println("Unable to update data: ", err)
		}
	}

	fmt.Println("finished MarkProcessed()", colName)
}

func DeleteExcessRows(pg *postgres, tableName string, metricName string, sortOrder string, maxRows int) {

	fmt.Println("Entered DeleteExcessRows() for ", tableName)
	fmt.Println("metricName: ", metricName)
	fmt.Println("sortOrder: ", sortOrder)
	fmt.Println("maxRows: ", maxRows)

	queryCount := `SELECT COUNT(*) FROM ` + tableName

	var rowCount int
	err := pg.db.QueryRow(context.Background(), queryCount).Scan(&rowCount)
	if err != nil {
		fmt.Println("Error querying db in DeleteExcessRows(): ", err)
		// return
	}

	fmt.Println("rowCount: ", rowCount)
	fmt.Println("maxRows: ", maxRows)

	if rowCount > maxRows {

		excessRows := rowCount - maxRows

		if excessRows <= 0 {
			fmt.Println("No excess rows in ", tableName)
			return
		}

		deleteStatement := `DELETE FROM ` + tableName + `
							WHERE id IN (
								SELECT id
								FROM ` + tableName + `
								ORDER BY ` + metricName + ` ` + sortOrder + ` , first_seen ASC
								LIMIT $1
								)`

		_, err := pg.db.Exec(context.Background(), deleteStatement, excessRows)
		if err != nil {
			fmt.Println("Failed to delete excess rows in ", tableName, ": ", err)
		}

		fmt.Printf("Deleted %d excess rows from %s, so it is back to %d rows", excessRows, tableName, maxRows)

	}
}
