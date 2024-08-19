package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func MarkProcessed(pg *postgres, colName string, aircrafts []Aircraft) {
	batch := &pgx.Batch{}

	for _, aircraft := range aircrafts {
		updateStatement := `UPDATE aircraft_data SET ` + colName + ` = true WHERE id = $1`
		batch.Queue(updateStatement, aircraft.Id)
	}

	_, err := pg.db.SendBatch(context.Background(), batch).Exec()

	if err != nil {
		fmt.Println("error: markProcessed() for colName: ", colName, ". Error: ", err)
		return
	}
}

func DeleteExcessRows(pg *postgres, tableName string, metricName string, sortOrder string, maxRows int) {

	fmt.Println("Entered DeleteExcessRows() for ", tableName)

	queryCount := `SELECT COUNT(*) FROM ` + tableName

	var rowCount int
	err := pg.db.QueryRow(context.Background(), queryCount).Scan(&rowCount)
	if err != nil {
		fmt.Println("Error querying db in DeleteExcessRows(): ", err)
		return
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

		_, err = pg.db.Exec(context.Background(), deleteStatement, excessRows)
		if err != nil {
			fmt.Println("Failed to delete excess rows in ", tableName, ": ", err)
		}

		fmt.Printf("Deleted %d excess rows from %s, so it is back to %d rows", excessRows, tableName, maxRows)

	}
}
