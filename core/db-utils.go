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

func GetLowestOrHighest(pg *postgres, tableName string, metricName string, ascOrDesc string) any {

	var returnValue any

	query := `SELECT ` + metricName +
		` FROM ` + tableName +
		` ORDER BY ` + metricName + ` ` + ascOrDesc +
		` LIMIT 1`

	err := pg.db.QueryRow(context.Background(), query).Scan(&returnValue)
	if err == nil {
		return returnValue
	} else {
		fmt.Println("GetLowestOrHighest() - No rows found in ", tableName, ". Default value will be used")
		return err
	}

}

func DeleteExcessRows(pg *postgres, tableName string, metricName string, sortOrder string, maxRows int) {

	queryCount := `SELECT COUNT(*) FROM ` + tableName

	var rowCount int
	err := pg.db.QueryRow(context.Background(), queryCount).Scan(&rowCount)
	if err != nil {
		fmt.Println("Error querying db in DeleteExcessRows(): ", err)
		return
	}

	if rowCount > maxRows {
		excessRows := rowCount - maxRows

		if excessRows <= 0 {
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
