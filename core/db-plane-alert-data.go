package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

type Row struct {
	ICAO         string
	Registration *string
	Operator     *string
	Type         *string
	ICAOType     *string
	Group        *string
	Tag1         *string
	Tag2         *string
	Tag3         *string
	Category     *string
	Link         *string
	Image1       *string
	Image2       *string
	Image3       *string
	Image4       *string
}

type GitHubAPIResponse struct {
	Files []struct {
		SHA      string `json:"sha"`
		Filename string `json:"name"`
	}
}

func UpsertPlaneAlertDb(pg *postgres) error {

	planeAlertUrl, isCustomPlaneAlertUrl := os.LookupEnv("PLANE_DB_URL")

	if !isCustomPlaneAlertUrl {
		planeAlertUrl = "https://raw.githubusercontent.com/sdr-enthusiasts/plane-alert-db/refs/heads/main/plane-alert-db-images.csv"
	}

	needsUpdating, commitHash, err := checkForUpdates(pg, isCustomPlaneAlertUrl)
	if err != nil {
		log.Warn().Msgf("Error checking for updates: %v", err)
		log.Warn().Msg("Updating despite error checking.")
		needsUpdating = true
		commitHash = "failed_to_get_commit_hash"
	}

	if !needsUpdating {
		log.Info().Msg("No new data in plane-alert-db, skipping update")
		return nil
	}

	log.Info().Msg("New data found in plane-alert-db, updating interesting aircraft reference data")

	planeAlertRecords, err := fetchCSVData(planeAlertUrl)
	if err != nil {
		return err
	}

	headers := getHeaderMap(planeAlertRecords[0])

	data := map[string]*Row{}

	for _, record := range planeAlertRecords[1:] {

		icao := record[headers["$ICAO"]]
		if icao == "" {
			continue
		}

		row := &Row{}
		row.ICAO = icao
		row.Registration = getValue(record[headers["$Registration"]])
		row.Operator = getValue(record[headers["$Operator"]])
		row.Type = getValue(record[headers["$Type"]])
		row.ICAOType = getValue(record[headers["$ICAO Type"]])
		row.Group = getValue(record[headers["#CMPG"]])
		row.Tag1 = getValue(record[headers["$Tag 1"]])
		row.Tag2 = getValue(record[headers["$#Tag 2"]])
		row.Tag3 = getValue(record[headers["$#Tag 3"]])
		row.Category = getValue(record[headers["Category"]])
		row.Link = getValue(record[headers["$#Link"]])
		row.Image1 = getValue(record[headers["#ImageLink"]])
		row.Image2 = getValue(record[headers["#ImageLink2"]])
		row.Image3 = getValue(record[headers["#ImageLink3"]])
		row.Image4 = getValue(record[headers["#ImageLink4"]])

		data[icao] = row
	}

	insertStatement := `
		INSERT INTO interesting_aircraft (
			icao, registration, operator, "type", icao_type,
			"group", tag1, tag2, tag3, category, link,
			image_link_1, image_link_2, image_link_3, image_link_4, commit_hash
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		)
		ON CONFLICT (icao) DO UPDATE SET
			registration = EXCLUDED.registration,
			operator = EXCLUDED.operator,
			"type" = EXCLUDED."type",
			icao_type = EXCLUDED.icao_type,
			"group" = EXCLUDED."group",
			tag1 = EXCLUDED.tag1,
			tag2 = EXCLUDED.tag2,
			tag3 = EXCLUDED.tag3,
			category = EXCLUDED.category,
			link = EXCLUDED.link,
			image_link_1 = EXCLUDED.image_link_1,
			image_link_2 = EXCLUDED.image_link_2,
			image_link_3 = EXCLUDED.image_link_3,
			image_link_4 = EXCLUDED.image_link_4,
			commit_hash = EXCLUDED.commit_hash
	`

	batch := &pgx.Batch{}
	for _, row := range data {
		batch.Queue(
			insertStatement,
			row.ICAO,
			row.Registration,
			row.Operator,
			row.Type,
			row.ICAOType,
			row.Group,
			row.Tag1,
			row.Tag2,
			row.Tag3,
			row.Category,
			row.Link,
			row.Image1,
			row.Image2,
			row.Image3,
			row.Image4,
			commitHash,
		)
	}

	br := pg.db.SendBatch(context.Background(), batch)
	defer br.Close()

	for range data {
		_, err := br.Exec()
		if err != nil {
			return fmt.Errorf("Error upserting interesting_aircraft data: %w", err)
		}
	}

	log.Info().Msgf("Succesfully upserted %d interesting aircraft records from plane-alert-db", len(data))

	return nil
}

func fetchCSVData(url string) ([][]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving the CSV for plane-alert-db: %w", err)
	}
	defer resp.Body.Close()

	reader := csv.NewReader(resp.Body)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("Error reading CSV records for plane-alert-db: %w", err)
	}

	return records, nil
}

func getHeaderMap(headers []string) map[string]int {
	headerMap := make(map[string]int)
	for i, header := range headers {
		headerMap[header] = i
	}
	return headerMap
}

func getValue(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func checkForUpdates(pg *postgres, isCustom bool) (needsUpdating bool, commitHash string, err error) {

	var exists bool
	err = pg.db.QueryRow(context.Background(), "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'interesting_aircraft')").Scan(&exists)
	if err != nil || !exists {
		return false, "", fmt.Errorf("Error checking for interesting_aircraft table: %w", err)
	}

	// If exists and is empty, then always needs updating
	var count int
	err = pg.db.QueryRow(context.Background(), "SELECT COUNT(*) FROM interesting_aircraft").Scan(&count)
	if err != nil {
		return false, "", fmt.Errorf("Error checking interesting_aircraft table: %w", err)
	}

	if count == 0 {
		return true, "", nil
	} else if isCustom { // If not empty and custom, skip updates, abort
		return false, "", nil
	}

	// Otherwise, check if newer commit hash
	var existingCommitHash sql.NullString
	err = pg.db.QueryRow(context.Background(), "SELECT commit_hash FROM interesting_aircraft LIMIT 1").Scan(&existingCommitHash)
	if err != nil {
		return false, "", fmt.Errorf("Error checking interesting_aircraft table: %w", err)
	}

	latestCommitHash, err := getLatestCommitHash()
	if err != nil {
		return false, "", fmt.Errorf("Error getting latest commit hash: %w", err)
	}

	if !existingCommitHash.Valid || latestCommitHash != existingCommitHash.String {
		return true, latestCommitHash, nil
	}

	return false, "", nil
}

func getLatestCommitHash() (string, error) {
	resp, err := http.Get("https://api.github.com/repos/sdr-enthusiasts/plane-alert-db/contents/")
	if err != nil {
		return "", fmt.Errorf("Error retrieving latest commit hash for plane-alert-db: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Error reading response body for latest commit hash for plane-alert-db: %w", err)
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Error when getting hash for plane-alert-db-images.csv: github api returned http code: %s", resp.Status)
	}

	var commitResponse GitHubAPIResponse
	err = json.Unmarshal(body, &commitResponse.Files)
	if err != nil {
		return "", fmt.Errorf("Error parsing JSON response for latest commit hash: %w", err)
	}

	for _, file := range commitResponse.Files {
		if file.Filename == "plane-alert-db-images.csv" {
			return file.SHA, nil
		}
	}
	log.Error().Msgf("getLatestCommitHash failed, printing commitResponse\n%+v\n", commitResponse)
	log.Error().Msgf("getLatestCommitHash failed, printing body\n%s\n", body)
	return "", fmt.Errorf("Error finding plane-alert-db-images.csv commit hash")
}
