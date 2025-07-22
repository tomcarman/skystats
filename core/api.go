package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type APIServer struct {
	pg   *postgres
	port string
}

func NewAPIServer(pg *postgres) *APIServer {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}
	return &APIServer{
		pg:   pg,
		port: port,
	}
}

func (s *APIServer) Start() {
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}

		c.Next()
	})

	// API routes (must come before static routes)
	api := r.Group("/api")
	{
		stats := api.Group("/stats")
		{
			stats.GET("/general", s.getGeneralStats)
			stats.GET("/above", s.getAboveStats)

			stats.GET("/routes/airlines", s.getTopAirlines)
			stats.GET("/routes/routes", s.getTopRoutes)
			stats.GET("/routes/countries-destination", s.getTopDestinationCountries)
			stats.GET("/routes/countries-origin", s.getTopOriginCountries)
			stats.GET("/routes/airports-domestic", s.getTopDomesticAirports)
			stats.GET("/routes/airports-international", s.getTopInternationalAirports)

			stats.GET("/motion/fastest", s.getFastestAircraft)
			stats.GET("/motion/slowest", s.getSlowestAircraft)
			stats.GET("/motion/highest", s.getHighestAircraft)
			stats.GET("/motion/lowest", s.getLowestAircraft)

			stats.GET("/interesting/civilian", s.getCivilianAircraft)
			stats.GET("/interesting/police", s.getPoliceAircraft)
			stats.GET("/interesting/military", s.getMilitaryAircraft)
			stats.GET("/interesting/government", s.getGovernmentAircraft)

		}
	}

	// Serve static files (must come after API routes)
	r.Static("/static", "../web")
	r.StaticFile("/", "../web/index.html")

	r.Run("0.0.0.0:" + s.port)
}

func (s *APIServer) getGeneralStats(c *gin.Context) {
	stats := gin.H{}

	// Total aircraft count
	var totalAircraft int
	err := s.pg.db.QueryRow(context.Background(), "SELECT COUNT(*) FROM aircraft_data").Scan(&totalAircraft)
	if err == nil {
		stats["total_aircraft"] = totalAircraft
	}

	// Today's aircraft count
	var todayAircraft int
	err = s.pg.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM aircraft_data WHERE DATE(first_seen) = CURRENT_DATE").Scan(&todayAircraft)
	if err == nil {
		stats["today_aircraft"] = todayAircraft
	}

	// Past hour aircraft count
	var hourAircraft int
	err = s.pg.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM aircraft_data WHERE first_seen >= NOW() - INTERVAL '1 hour'").Scan(&hourAircraft)
	if err == nil {
		stats["hour_aircraft"] = hourAircraft
	}

	// Unique aircraft types
	var uniqueTypes int
	err = s.pg.db.QueryRow(context.Background(),
		"SELECT COUNT(DISTINCT t) FROM aircraft_data WHERE t IS NOT NULL AND t != ''").Scan(&uniqueTypes)
	if err == nil {
		stats["unique_aircraft_types"] = uniqueTypes
	}

	// Interesting aircraft count
	var interestingCount int
	err = s.pg.db.QueryRow(context.Background(), "SELECT COUNT(*) FROM interesting_aircraft_seen").Scan(&interestingCount)
	if err == nil {
		stats["interesting_aircraft_count"] = interestingCount
	}

	c.JSON(http.StatusOK, stats)
}

func (s *APIServer) getAboveStats(c *gin.Context) {

	radiusValue := os.Getenv("ABOVE_RADIUS")
	radius, err := strconv.Atoi(radiusValue)
	if err != nil || radius <= 0 {
		fmt.Println("Error parsing ABOVE_RADIUS environment variable ", err)
		return
	}

	query := `
		SELECT 
			ad.hex, 
			ad.flight, 
			ad.r, 
			ad.t, 
			ad.track, 
			ad.first_seen, 
			ad.last_seen,
			ad.last_seen_lat, 
			ad.last_seen_lon, 
			ad.last_seen_distance,
			-- Registration data
			reg.type,
			reg.icao_type,
			reg.manufacturer,
			reg.registered_owner_country_name,
			reg.registered_owner_country_iso_name,
			reg.registered_owner_operator_flag_code,
			reg.registered_owner,
			reg.url_photo,
			reg.url_photo_thumbnail,
			-- Route data
			rt.airline_name,
			rt.airline_icao,
			rt.origin_country_name,
			rt.origin_country_iso_name,
			rt.origin_iata_code,
			rt.origin_icao_code,
			rt.origin_name,
			rt.destination_country_name,
			rt.destination_country_iso_name,
			rt.destination_iata_code,
			rt.destination_icao_code,
			rt.destination_name
		FROM aircraft_data ad
		LEFT JOIN registration_data reg ON ad.hex = reg.mode_s
		LEFT JOIN route_data rt ON ad.flight = rt.route_callsign
		WHERE ad.last_seen >= NOW() - INTERVAL '60 seconds'
			AND ad.last_seen_distance <= $1
		ORDER BY ad.last_seen_distance ASC
		LIMIT 5;`

	rows, err := s.pg.db.Query(context.Background(), query, radius)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	aircraft := []gin.H{}
	for rows.Next() {
		// Core data
		var hex, flight, registration, aircraftType string
		var firstSeen, lastSeen *time.Time
		var track, lastSeenLat, lastSeenLon, lastSeenDistance float64

		// Registration data
		var regType, icaoType, manufacturer, registeredOwnerCountryName, registeredOwnerCountryISO, registeredOwnerOperatorFlag, registeredOwner *string
		var urlPhoto, urlPhotoThumbnail *string

		// Route data
		var airlineName, airlineICAO, originCountryName, originCountryISOName, originIATACode, originICAOCode, originName *string
		var destinationCountryName, destinationCountryISOName, destinationIATACode, destinationICAOCode, destinationName *string

		err := rows.Scan(
			// Core data
			&hex, &flight, &registration, &aircraftType, &track,
			&firstSeen, &lastSeen, &lastSeenLat, &lastSeenLon, &lastSeenDistance,

			// Registration data
			&regType, &icaoType, &manufacturer, &registeredOwnerCountryName, &registeredOwnerCountryISO,
			&registeredOwnerOperatorFlag, &registeredOwner, &urlPhoto, &urlPhotoThumbnail,

			// Route data
			&airlineName, &airlineICAO, &originCountryName, &originCountryISOName, &originIATACode,
			&originICAOCode, &originName, &destinationCountryName, &destinationCountryISOName,
			&destinationIATACode, &destinationICAOCode, &destinationName)
		if err != nil {
			fmt.Println("Error in getAboveStats() ", err)
			continue
		}

		aircraft = append(aircraft, gin.H{
			// Core data
			"hex":                hex,
			"flight":             flight,
			"registration":       registration,
			"type":               aircraftType,
			"first_seen":         firstSeen,
			"last_seen":          lastSeen,
			"last_seen_lat":      lastSeenLat,
			"last_seen_lon":      lastSeenLon,
			"last_seen_distance": lastSeenDistance,
			"track":              track,
			// Registration data
			"reg_type":                       regType,
			"icao_type":                      icaoType,
			"manufacturer":                   manufacturer,
			"registered_owner_country_name":  registeredOwnerCountryName,
			"registered_owner_country_iso":   registeredOwnerCountryISO,
			"registered_owner_operator_flag": registeredOwnerOperatorFlag,
			"registered_owner":               registeredOwner,
			"url_photo":                      urlPhoto,
			"url_photo_thumbnail":            urlPhotoThumbnail,
			// Route data
			"airline_name":                 airlineName,
			"airline_icao":                 airlineICAO,
			"origin_country_name":          originCountryName,
			"origin_country_iso_name":      originCountryISOName,
			"origin_iata_code":             originIATACode,
			"origin_icao_code":             originICAOCode,
			"origin_name":                  originName,
			"destination_country_name":     destinationCountryName,
			"destination_country_iso_name": destinationCountryISOName,
			"destination_iata_code":        destinationIATACode,
			"destination_icao_code":        destinationICAOCode,
			"destination_name":             destinationName,
		})
	}

	c.JSON(http.StatusOK, aircraft)
}

func (s *APIServer) getGroupedInterestingAircraft(c *gin.Context, query string, limit int) {
	rows, err := s.pg.db.Query(context.Background(), query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	aircraft := []gin.H{}
	for rows.Next() {
		var icao, registration, operator, aircraftType, icaoType, group, category string
		var tag1, tag2, tag3 string
		var imageLink1, imageLink2, imageLink3 *string // Use pointers for nullable fields
		var hex, flight string
		var seen *time.Time
		var seenEpoch float64

		err := rows.Scan(&icao, &registration, &operator, &aircraftType, &icaoType,
			&group, &category, &tag1, &tag2, &tag3, &imageLink1, &imageLink2, &imageLink3,
			&hex, &flight, &seen, &seenEpoch)
		if err != nil {
			continue
		}

		aircraft = append(aircraft, gin.H{
			"icao":         icao,
			"registration": registration,
			"operator":     operator,
			"type":         aircraftType,
			"icao_type":    icaoType,
			"group":        group,
			"category":     category,
			"tag1":         tag1,
			"tag2":         tag2,
			"tag3":         tag3,
			"image_link_1": imageLink1,
			"image_link_2": imageLink2,
			"image_link_3": imageLink3,
			"hex":          hex,
			"flight":       flight,
			"seen":         seen,
			"seen_epoch":   seenEpoch,
		})
	}

	c.JSON(http.StatusOK, aircraft)
}

func (s *APIServer) getCivilianAircraft(c *gin.Context) {
	limit := s.getLimit(c)

	query := `
		SELECT icao, registration, operator, type, icao_type, "group", 
			   category, tag1, tag2, tag3, image_link_1, image_link_2, image_link_3,
			   hex, flight, seen, seen_epoch
		FROM interesting_aircraft_seen 
		WHERE "group" = 'Civ'
		ORDER BY seen DESC 
		LIMIT $1`

	s.getGroupedInterestingAircraft(c, query, limit)
}

func (s *APIServer) getPoliceAircraft(c *gin.Context) {
	limit := s.getLimit(c)

	query := `
		SELECT icao, registration, operator, type, icao_type, "group", 
			   category, tag1, tag2, tag3, image_link_1, image_link_2, image_link_3,
			   hex, flight, seen, seen_epoch
		FROM interesting_aircraft_seen 
		WHERE "group" = 'Pol'
		ORDER BY seen DESC 
		LIMIT $1`

	s.getGroupedInterestingAircraft(c, query, limit)
}

func (s *APIServer) getMilitaryAircraft(c *gin.Context) {
	limit := s.getLimit(c)

	query := `
		SELECT icao, registration, operator, type, icao_type, "group", 
			   category, tag1, tag2, tag3, image_link_1, image_link_2, image_link_3,
			   hex, flight, seen, seen_epoch
		FROM interesting_aircraft_seen 
		WHERE "group" = 'Mil'
		ORDER BY seen DESC 
		LIMIT $1`

	s.getGroupedInterestingAircraft(c, query, limit)
}

func (s *APIServer) getGovernmentAircraft(c *gin.Context) {
	limit := s.getLimit(c)

	query := `
		SELECT icao, registration, operator, type, icao_type, "group", 
			   category, tag1, tag2, tag3, image_link_1, image_link_2, image_link_3,
			   hex, flight, seen, seen_epoch
		FROM interesting_aircraft_seen 
		WHERE "group" = 'Gov'
		ORDER BY seen DESC 
		LIMIT $1`

	s.getGroupedInterestingAircraft(c, query, limit)
}

func (s *APIServer) getFastestAircraft(c *gin.Context) {
	limit := s.getLimit(c)

	query := `
		SELECT hex, flight, registration, type, first_seen, last_seen, 
			   ground_speed, indicated_air_speed, true_air_speed
		FROM fastest_aircraft 
		ORDER BY ground_speed DESC 
		LIMIT $1`

	rows, err := s.pg.db.Query(context.Background(), query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	aircraft := []gin.H{}
	for rows.Next() {
		var hex, flight, registration, aircraftType string
		var firstSeen, lastSeen *time.Time
		var groundSpeed float64
		var indicatedAirSpeed, trueAirSpeed int

		err := rows.Scan(&hex, &flight, &registration, &aircraftType, &firstSeen,
			&lastSeen, &groundSpeed, &indicatedAirSpeed, &trueAirSpeed)
		if err != nil {
			continue
		}

		aircraft = append(aircraft, gin.H{
			"hex":                 hex,
			"flight":              flight,
			"registration":        registration,
			"type":                aircraftType,
			"first_seen":          firstSeen,
			"last_seen":           lastSeen,
			"ground_speed":        groundSpeed,
			"indicated_air_speed": indicatedAirSpeed,
			"true_air_speed":      trueAirSpeed,
		})
	}

	c.JSON(http.StatusOK, aircraft)
}

func (s *APIServer) getSlowestAircraft(c *gin.Context) {
	limit := s.getLimit(c)

	query := `
		SELECT hex, flight, registration, type, first_seen, last_seen, 
			   ground_speed, indicated_air_speed, true_air_speed
		FROM slowest_aircraft 
		ORDER BY ground_speed ASC 
		LIMIT $1`

	rows, err := s.pg.db.Query(context.Background(), query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	aircraft := []gin.H{}
	for rows.Next() {
		var hex, flight, registration, aircraftType string
		var firstSeen, lastSeen *time.Time
		var groundSpeed float64
		var indicatedAirSpeed, trueAirSpeed int

		err := rows.Scan(&hex, &flight, &registration, &aircraftType, &firstSeen,
			&lastSeen, &groundSpeed, &indicatedAirSpeed, &trueAirSpeed)
		if err != nil {
			continue
		}

		aircraft = append(aircraft, gin.H{
			"hex":                 hex,
			"flight":              flight,
			"registration":        registration,
			"type":                aircraftType,
			"first_seen":          firstSeen,
			"last_seen":           lastSeen,
			"ground_speed":        groundSpeed,
			"indicated_air_speed": indicatedAirSpeed,
			"true_air_speed":      trueAirSpeed,
		})
	}

	c.JSON(http.StatusOK, aircraft)
}

func (s *APIServer) getHighestAircraft(c *gin.Context) {
	limit := s.getLimit(c)

	query := `
		SELECT hex, flight, registration, type, first_seen, last_seen, 
			   barometric_altitude, geometric_altitude
		FROM highest_aircraft 
		ORDER BY barometric_altitude DESC 
		LIMIT $1`

	rows, err := s.pg.db.Query(context.Background(), query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	aircraft := []gin.H{}
	for rows.Next() {
		var hex, flight, registration, aircraftType string
		var firstSeen, lastSeen *time.Time
		var barometricAltitude, geometricAltitude int

		err := rows.Scan(&hex, &flight, &registration, &aircraftType, &firstSeen,
			&lastSeen, &barometricAltitude, &geometricAltitude)
		if err != nil {
			continue
		}

		aircraft = append(aircraft, gin.H{
			"hex":                 hex,
			"flight":              flight,
			"registration":        registration,
			"type":                aircraftType,
			"first_seen":          firstSeen,
			"last_seen":           lastSeen,
			"barometric_altitude": barometricAltitude,
			"geometric_altitude":  geometricAltitude,
		})
	}

	c.JSON(http.StatusOK, aircraft)
}

func (s *APIServer) getLowestAircraft(c *gin.Context) {
	limit := s.getLimit(c)

	query := `
		SELECT hex, flight, registration, type, first_seen, last_seen, 
			   barometric_altitude, geometric_altitude
		FROM lowest_aircraft 
		ORDER BY barometric_altitude ASC 
		LIMIT $1`

	rows, err := s.pg.db.Query(context.Background(), query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	aircraft := []gin.H{}
	for rows.Next() {
		var hex, flight, registration, aircraftType string
		var firstSeen, lastSeen *time.Time
		var barometricAltitude, geometricAltitude int

		err := rows.Scan(&hex, &flight, &registration, &aircraftType, &firstSeen,
			&lastSeen, &barometricAltitude, &geometricAltitude)
		if err != nil {
			continue
		}

		aircraft = append(aircraft, gin.H{
			"hex":                 hex,
			"flight":              flight,
			"registration":        registration,
			"type":                aircraftType,
			"first_seen":          firstSeen,
			"last_seen":           lastSeen,
			"barometric_altitude": barometricAltitude,
			"geometric_altitude":  geometricAltitude,
		})
	}

	c.JSON(http.StatusOK, aircraft)
}

func (s *APIServer) getTopRoutes(c *gin.Context) {
	limit := s.getLimit(c)

	query := `
		SELECT 
			CONCAT(rd.origin_iata_code, ' → ', rd.destination_iata_code) as route,
			rd.origin_iata_code,
			rd.origin_name,
			rd.destination_iata_code,
			rd.destination_name,
			COUNT(*) as flight_count
		FROM aircraft_data ad 
		INNER JOIN route_data rd ON ad.flight = rd.route_callsign
		WHERE rd.origin_iata_code IS NOT NULL AND rd.origin_iata_code != ''
			AND rd.destination_iata_code IS NOT NULL AND rd.destination_iata_code != ''
			AND rd.origin_iata_code != rd.destination_iata_code
		GROUP BY rd.origin_iata_code, rd.origin_name, rd.destination_iata_code, rd.destination_name
		ORDER BY flight_count DESC
		LIMIT $1`

	rows, err := s.pg.db.Query(context.Background(), query, limit)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	results := []gin.H{}

	for rows.Next() {
		var route, origin_iata_code, origin_name, destination_iata_code, destination_name string
		var flight_count int

		err := rows.Scan(&route, &origin_iata_code, &origin_name, &destination_iata_code, &destination_name, &flight_count)
		if err != nil {
			continue
		}

		results = append(results, gin.H{
			"route":                 route,
			"origin_iata_code":      origin_iata_code,
			"origin_name":           origin_name,
			"destination_iata_code": destination_iata_code,
			"destination_name":      destination_name,
			"flight_count":          flight_count,
		})
	}

	c.JSON(http.StatusOK, results)

}

func (s *APIServer) getTopDestinationCountries(c *gin.Context) {
	limit := s.getLimit(c)

	query := `
		SELECT 
			rd.destination_country_name,
			rd.destination_country_iso_name,
			COUNT(*) as flight_count
		FROM aircraft_data ad 
		INNER JOIN route_data rd ON ad.flight = rd.route_callsign
		WHERE rd.destination_country_iso_name IS NOT NULL AND rd.destination_country_iso_name != ''
			AND rd.origin_country_iso_name != rd.destination_country_iso_name
		GROUP BY rd.destination_country_name, destination_country_iso_name
		ORDER BY flight_count DESC
		LIMIT $1`

	rows, err := s.pg.db.Query(context.Background(), query, limit)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	results := []gin.H{}

	for rows.Next() {
		var destination_country_name, destination_country_iso_name string
		var flight_count int

		err := rows.Scan(&destination_country_name, &destination_country_iso_name, &flight_count)
		if err != nil {
			continue
		}

		results = append(results, gin.H{
			"country_name": destination_country_name,
			"country_iso":  destination_country_iso_name,
			"flight_count": flight_count,
		})
	}

	c.JSON(http.StatusOK, results)

}

func (s *APIServer) getTopOriginCountries(c *gin.Context) {
	limit := s.getLimit(c)

	query := `
		SELECT 
			rd.origin_country_name,
			rd.origin_country_iso_name,
			COUNT(*) as flight_count
		FROM aircraft_data ad 
		INNER JOIN route_data rd ON ad.flight = rd.route_callsign
		WHERE rd.origin_country_iso_name IS NOT NULL AND rd.origin_country_iso_name != ''
			AND rd.destination_country_iso_name != rd.origin_country_iso_name
		GROUP BY rd.origin_country_name, origin_country_iso_name
		ORDER BY flight_count DESC
		LIMIT $1`

	rows, err := s.pg.db.Query(context.Background(), query, limit)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	results := []gin.H{}

	for rows.Next() {
		var origin_country_name, origin_country_iso_name string
		var flight_count int

		err := rows.Scan(&origin_country_name, &origin_country_iso_name, &flight_count)
		if err != nil {
			continue
		}

		results = append(results, gin.H{
			"country_name": origin_country_name,
			"country_iso":  origin_country_iso_name,
			"flight_count": flight_count,
		})
	}

	c.JSON(http.StatusOK, results)

}

func (s *APIServer) getTopAirlines(c *gin.Context) {
	limit := s.getLimit(c)

	query := `
		SELECT 
			rd.airline_name,
			rd.airline_icao,
			rd.airline_iata,
			COUNT(*) as flight_count
		FROM aircraft_data ad 
		INNER JOIN route_data rd ON ad.flight = rd.route_callsign
		WHERE rd.airline_name IS NOT NULL AND rd.airline_name != ''
			AND rd.origin_iata_code != rd.destination_iata_code
			AND rd.origin_iata_code IS NOT NULL AND rd.origin_iata_code != ''
			AND rd.destination_iata_code IS NOT NULL AND rd.destination_iata_code != ''
		GROUP BY rd.airline_name, rd.airline_icao, rd.airline_iata
		ORDER BY flight_count DESC
		LIMIT $1`

	rows, err := s.pg.db.Query(context.Background(), query, limit)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	results := []gin.H{}

	for rows.Next() {
		var airline_name, airline_icao, airline_iata string
		var flight_count int

		err := rows.Scan(&airline_name, &airline_icao, &airline_iata, &flight_count)
		if err != nil {
			continue
		}

		results = append(results, gin.H{
			"airline_name": airline_name,
			"airline_icao": airline_icao,
			"airline_iata": airline_iata,
			"flight_count": flight_count,
		})
	}

	c.JSON(http.StatusOK, results)

}

func (s *APIServer) getTopDomesticAirports(c *gin.Context) {
	limit := s.getLimit(c)

	query := `
		SELECT
			airport_code,
			airport_name,
			airport_country,
			SUM(flight_count) as flight_count
		FROM (
			SELECT
				rd.origin_iata_code as airport_code,
				rd.origin_name as airport_name,
				rd.origin_country_name as airport_country,
				COUNT(*) as flight_count
			FROM aircraft_data ad
			INNER JOIN route_data rd ON ad.flight = rd.route_callsign
			WHERE rd.origin_country_iso_name = $1
				AND rd.origin_iata_code IS NOT NULL AND rd.origin_iata_code != ''
				AND rd.destination_iata_code IS NOT NULL AND rd.destination_iata_code != ''
				AND rd.origin_iata_code != rd.destination_iata_code
			GROUP BY rd.origin_iata_code, rd.origin_name, rd.origin_country_name
			UNION ALL
			SELECT
				rd.destination_iata_code as airport_code,
				rd.destination_name as airport_name,
				rd.destination_country_name as airport_country,
				COUNT(*) as flight_count
			FROM aircraft_data ad
			INNER JOIN route_data rd ON ad.flight = rd.route_callsign
			WHERE rd.destination_country_iso_name = $1
				AND rd.origin_iata_code IS NOT NULL AND rd.origin_iata_code != ''
				AND rd.destination_iata_code IS NOT NULL AND rd.destination_iata_code != ''
				AND rd.origin_iata_code != rd.destination_iata_code
			GROUP BY rd.destination_iata_code, rd.destination_name, rd.destination_country_name
		) combined_airports
		GROUP BY airport_code, airport_name, airport_country
		ORDER BY flight_count DESC
		LIMIT $2`

	rows, err := s.pg.db.Query(context.Background(), query, s.getCountry(), limit)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	results := []gin.H{}

	for rows.Next() {
		var airport_code, airport_name, airport_country string
		var flight_count int

		err := rows.Scan(&airport_code, &airport_name, &airport_country, &flight_count)
		if err != nil {
			continue
		}

		results = append(results, gin.H{
			"airport_code":    airport_code,
			"airport_name":    airport_name,
			"airport_country": airport_country,
			"flight_count":    flight_count,
		})
	}

	c.JSON(http.StatusOK, results)

}

func (s *APIServer) getTopInternationalAirports(c *gin.Context) {
	limit := s.getLimit(c)

	fmt.Printf("GET COUNTRY: %s\n", s.getCountry())

	query := `
		SELECT
			airport_code,
			airport_name,
			airport_country,
			SUM(flight_count) as flight_count
		FROM (
			SELECT
				rd.origin_iata_code as airport_code,
				rd.origin_name as airport_name,
				rd.origin_country_name as airport_country,
				COUNT(*) as flight_count
			FROM aircraft_data ad
			INNER JOIN route_data rd ON ad.flight = rd.route_callsign
			WHERE rd.origin_country_iso_name != $1
				AND rd.origin_iata_code IS NOT NULL AND rd.origin_iata_code != ''
				AND rd.destination_iata_code IS NOT NULL AND rd.destination_iata_code != ''
				AND rd.origin_iata_code != rd.destination_iata_code
			GROUP BY rd.origin_iata_code, rd.origin_name, rd.origin_country_name
			UNION ALL
			SELECT
				rd.destination_iata_code as airport_code,
				rd.destination_name as airport_name,
				rd.destination_country_name as airport_country,
				COUNT(*) as flight_count
			FROM aircraft_data ad
			INNER JOIN route_data rd ON ad.flight = rd.route_callsign
			WHERE rd.destination_country_iso_name != $1
				AND rd.origin_iata_code IS NOT NULL AND rd.origin_iata_code != ''
				AND rd.destination_iata_code IS NOT NULL AND rd.destination_iata_code != ''
				AND rd.origin_iata_code != rd.destination_iata_code
			GROUP BY rd.destination_iata_code, rd.destination_name, rd.destination_country_name
		) combined_airports
		GROUP BY airport_code, airport_name, airport_country
		ORDER BY flight_count DESC
		LIMIT $2`

	rows, err := s.pg.db.Query(context.Background(), query, s.getCountry(), limit)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	results := []gin.H{}

	for rows.Next() {
		var airport_code, airport_name, airport_country string
		var flight_count int

		err := rows.Scan(&airport_code, &airport_name, &airport_country, &flight_count)
		if err != nil {
			continue
		}

		results = append(results, gin.H{
			"airport_code":    airport_code,
			"airport_name":    airport_name,
			"airport_country": airport_country,
			"flight_count":    flight_count,
		})
	}

	c.JSON(http.StatusOK, results)

}

func (s *APIServer) getLimit(c *gin.Context) int {
	limitStr := c.DefaultQuery("limit", "5")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		return 10
	}

	if limit > 100 {
		return 100 // max limit
	}

	return limit
}

func (s *APIServer) getCountry() string {
	return os.Getenv("DOMESTIC_COUNTRY_ISO")
}
