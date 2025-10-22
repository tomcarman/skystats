package main

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

func MarkProcessed(pg *postgres, colName string, aircrafts []Aircraft) {

	batch := &pgx.Batch{}

	for _, aircraft := range aircrafts {
		updateStatement := `UPDATE aircraft_data SET ` + colName + ` = true WHERE id = $1`
		batch.Queue(updateStatement, aircraft.Id)
	}

	br := pg.db.SendBatch(context.Background(), batch)
	defer br.Close()

	for i := 0; i < len(aircrafts); i++ {
		_, err := br.Exec()
		if err != nil {
			log.Error().Err(err).Msg("MarkProcessed() - Unable to update data")
		}
	}
}

func DeleteExcessRows(pg *postgres, tableName string, metricName string, sortOrder string, maxRows int) {

	queryCount := `SELECT COUNT(*) FROM ` + tableName

	var rowCount int
	err := pg.db.QueryRow(context.Background(), queryCount).Scan(&rowCount)
	if err != nil {
		log.Error().Err(err).Msg("DeleteExcessRows() - Error querying db in DeleteExcessRows()")
		return
	}

	if rowCount > maxRows {

		excessRows := rowCount - maxRows

		if excessRows <= 0 {
			log.Debug().Msgf("DeleteExcessRows() - No excess rows in %s", tableName)
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
			log.Error().Err(err).Msgf("DeleteExcessRows() - Failed to delete excess rows in %s", tableName)
		}
	}
}
