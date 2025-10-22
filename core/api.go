package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

type APIServer struct {
	pg       *postgres
	port     string
	settings *SettingsService
}

func NewAPIServer(pg *postgres) *APIServer {
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}
	return &APIServer{
		pg:       pg,
		port:     port,
		settings: NewSettingsService(pg),
	}
}

func (s *APIServer) getTimezone(c *gin.Context) string {

	tz := c.Query("tz")
	if tz == "" {
		return "UTC"
	}

	_, err := time.LoadLocation(tz)
	if err != nil {
		return "UTC"
	}

	return tz
}

func (s *APIServer) Start() {

	// Start gin API server in release or debug mode based on LOG_LEVEL
	logLevel := os.Getenv("LOG_LEVEL")
	var r *gin.Engine
	if logLevel == "DEBUG" || logLevel == "TRACE" {
		r = gin.Default()
	} else {
		gin.SetMode(gin.ReleaseMode)
		r = gin.New()
		r.Use(gin.Recovery())
	}

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

	// API routes
	api := r.Group("api")
	{
		stats := api.Group("/stats")
		{
			stats.GET("/above", s.getAboveStats)

			stats.GET("/seen/flights", s.getFlightsSeenMetrics)
			stats.GET("/seen/aircraft", s.getAircraftSeenMetrics)

			stats.GET("/routes/metrics", s.getRouteMetrics)
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

			stats.GET("/interesting/metrics", s.getInterestingMetrics)
			stats.GET("/interesting/civilian", func(c *gin.Context) { s.getRecentInterestingAircraft(c, "Civ") })
			stats.GET("/interesting/police", func(c *gin.Context) { s.getRecentInterestingAircraft(c, "Pol") })
			stats.GET("/interesting/military", func(c *gin.Context) { s.getRecentInterestingAircraft(c, "Mil") })
			stats.GET("/interesting/government", func(c *gin.Context) { s.getRecentInterestingAircraft(c, "Gov") })

			stats.GET("/types/flights/all", func(c *gin.Context) { s.getTopAircraftTypes(c, "all", "flights") })
			stats.GET("/types/flights/year", func(c *gin.Context) { s.getTopAircraftTypes(c, "year", "flights") })
			stats.GET("/types/flights/month", func(c *gin.Context) { s.getTopAircraftTypes(c, "month", "flights") })
			stats.GET("/types/flights/day", func(c *gin.Context) { s.getTopAircraftTypes(c, "day", "flights") })

			stats.GET("/types/aircraft/all", func(c *gin.Context) { s.getTopAircraftTypes(c, "all", "aircraft") })
			stats.GET("/types/aircraft/year", func(c *gin.Context) { s.getTopAircraftTypes(c, "year", "aircraft") })
			stats.GET("/types/aircraft/month", func(c *gin.Context) { s.getTopAircraftTypes(c, "month", "aircraft") })
			stats.GET("/types/aircraft/day", func(c *gin.Context) { s.getTopAircraftTypes(c, "day", "aircraft") })

			stats.GET("/charts/flights/year", func(c *gin.Context) { s.getChartFlightsOverTime(c, "year") })
			stats.GET("/charts/flights/month", func(c *gin.Context) { s.getChartFlightsOverTime(c, "month") })
			stats.GET("/charts/flights/day", func(c *gin.Context) { s.getChartFlightsOverTime(c, "day") })

			stats.GET("/charts/aircraft/year", func(c *gin.Context) { s.getChartAircraftOverTime(c, "year") })
			stats.GET("/charts/aircraft/month", func(c *gin.Context) { s.getChartAircraftOverTime(c, "month") })
			stats.GET("/charts/aircraft/day", func(c *gin.Context) { s.getChartAircraftOverTime(c, "day") })

		}

		settings := api.Group("/settings")
		{
			settings.GET("", s.getSettings)
			settings.PUT("", s.updateSettings)
		}

		api.GET("/version", s.getVersion)
	}

	// Serve static files
	r.Static("/static", "../web")
	r.StaticFile("/", "../web/index.html")

	r.Run("0.0.0.0:" + s.port)

}
func (s *APIServer) getFlightsSeenMetrics(c *gin.Context) {
	stats := gin.H{}

	tz := s.getTimezone(c)

	// Total flights count
	var totalFlights int
	err := s.pg.db.QueryRow(context.Background(), "SELECT COUNT(*) FROM aircraft_data").Scan(&totalFlights)
	if err == nil {
		stats["total_flights"] = totalFlights
	}

	// Today's flights count
	var todayFlights int
	err = s.pg.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM aircraft_data WHERE DATE(first_seen AT TIME ZONE $1) = CURRENT_DATE", tz).Scan(&todayFlights)
	if err == nil {
		stats["today_flights"] = todayFlights
	}

	// Past hour flights count
	var hourFlights int
	err = s.pg.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM aircraft_data WHERE first_seen >= NOW() - INTERVAL '1 hour'").Scan(&hourFlights)
	if err == nil {
		stats["hour_flights"] = hourFlights
	}

	c.JSON(http.StatusOK, stats)

}

func (s *APIServer) getAircraftSeenMetrics(c *gin.Context) {
	stats := gin.H{}

	tz := s.getTimezone(c)

	// Total aircraft count
	var totalAircraft int
	err := s.pg.db.QueryRow(context.Background(), "SELECT COUNT(DISTINCT hex) FROM aircraft_data").Scan(&totalAircraft)
	if err == nil {
		stats["total_aircraft"] = totalAircraft
	}

	// Today's aircraft count
	var todayAircraft int
	err = s.pg.db.QueryRow(context.Background(),
		"SELECT COUNT(DISTINCT hex) FROM aircraft_data WHERE DATE(first_seen AT TIME ZONE $1) = CURRENT_DATE", tz).Scan(&todayAircraft)
	if err == nil {
		stats["today_aircraft"] = todayAircraft
	}

	// Past hour aircraft count
	var hourAircraft int
	err = s.pg.db.QueryRow(context.Background(),
		"SELECT COUNT(DISTINCT hex) FROM aircraft_data WHERE first_seen >= NOW() - INTERVAL '1 hour'").Scan(&hourAircraft)
	if err == nil {
		stats["hour_aircraft"] = hourAircraft
	}

	c.JSON(http.StatusOK, stats)
}

func (s *APIServer) getRouteMetrics(c *gin.Context) {

	stats := gin.H{}

	// Total Routes
	var total_routes int
	err := s.pg.db.QueryRow(context.Background(),
		`SELECT COUNT(*)
			FROM aircraft_data a
			INNER JOIN route_data r ON a.flight = r.route_callsign`).Scan(&total_routes)

	if err == nil {
		stats["total_routes"] = total_routes
	}

	// Unique countries
	var uniqueCountries int
	err = s.pg.db.QueryRow(context.Background(),
		`SELECT COUNT(*) 
		FROM (
			SELECT origin_country_name AS country FROM route_data
			UNION 
			SELECT destination_country_name AS country FROM route_data
		) AS unique_countries`).Scan(&uniqueCountries)

	if err == nil {
		stats["unqiue_countries"] = uniqueCountries
	}

	// Unique airports
	var uniqueAirports int
	err = s.pg.db.QueryRow(context.Background(),
		`SELECT COUNT(*) 
		FROM (
			SELECT origin_icao_code AS airport FROM route_data
			UNION 
			SELECT destination_icao_code AS airport FROM route_data
		) AS unique_airports`).Scan(&uniqueAirports)

	if err == nil {
		stats["unique_airports"] = uniqueAirports
	}

	c.JSON(http.StatusOK, stats)

}

func (s *APIServer) getInterestingMetrics(c *gin.Context) {
	stats := gin.H{}

	tz := s.getTimezone(c)

	// Interesting aircraft count
	var interestingCount int
	err := s.pg.db.QueryRow(context.Background(), "SELECT COUNT(*) FROM interesting_aircraft_seen").Scan(&interestingCount)
	if err == nil {
		stats["total_interesting"] = interestingCount
	}

	// Today's interesting aircraft count
	var todayInterestingCount int
	err = s.pg.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM interesting_aircraft_seen WHERE DATE(first_seen AT TIME ZONE $1) = CURRENT_DATE", tz).Scan(&todayInterestingCount)
	if err == nil {
		stats["today_interesting"] = todayInterestingCount
	}

	// Past hour interesting aircraft count
	var hourInterestingCount int
	err = s.pg.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM interesting_aircraft_seen WHERE first_seen >= NOW() - INTERVAL '1 hour'").Scan(&hourInterestingCount)
	if err == nil {
		stats["hour_interesting"] = hourInterestingCount
	}

	c.JSON(http.StatusOK, stats)

}

func (s *APIServer) getAboveStats(c *gin.Context) {

	radiusValue := os.Getenv("ABOVE_RADIUS")
	radius, err := strconv.Atoi(radiusValue)
	if err != nil || radius <= 0 {
		log.Error().Err(err).Msg("Error parsing ABOVE_RADIUS environment variable")
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
			ad.destination_distance,
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
			rt.destination_name,
			rt.route_distance
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
		var destinationDistance *float64

		// Registration data
		var regType, icaoType, manufacturer, registeredOwnerCountryName, registeredOwnerCountryISO, registeredOwnerOperatorFlag, registeredOwner *string
		var urlPhoto, urlPhotoThumbnail *string

		// Route data
		var airlineName, airlineICAO, originCountryName, originCountryISOName, originIATACode, originICAOCode, originName *string
		var destinationCountryName, destinationCountryISOName, destinationIATACode, destinationICAOCode, destinationName *string
		var routeDistance *float64

		err := rows.Scan(
			// Core data
			&hex, &flight, &registration, &aircraftType, &track,
			&firstSeen, &lastSeen, &lastSeenLat, &lastSeenLon, &lastSeenDistance, &destinationDistance,

			// Registration data
			&regType, &icaoType, &manufacturer, &registeredOwnerCountryName, &registeredOwnerCountryISO,
			&registeredOwnerOperatorFlag, &registeredOwner, &urlPhoto, &urlPhotoThumbnail,

			// Route data
			&airlineName, &airlineICAO, &originCountryName, &originCountryISOName, &originIATACode,
			&originICAOCode, &originName, &destinationCountryName, &destinationCountryISOName,
			&destinationIATACode, &destinationICAOCode, &destinationName, &routeDistance)
		if err != nil {
			log.Error().Err(err).Msg("getAboveStats()")
			continue
		}

		aircraft = append(aircraft, gin.H{
			// Core data
			"hex":                  hex,
			"flight":               flight,
			"registration":         registration,
			"type":                 aircraftType,
			"first_seen":           firstSeen,
			"last_seen":            lastSeen,
			"last_seen_lat":        lastSeenLat,
			"last_seen_lon":        lastSeenLon,
			"last_seen_distance":   lastSeenDistance,
			"destination_distance": destinationDistance,
			"track":                track,
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
			"route_distance":               routeDistance,
		})
	}

	c.JSON(http.StatusOK, aircraft)
}

func (s *APIServer) getRecentInterestingAircraft(c *gin.Context, group string) {

	limit := s.getLimit("interesting_table_limit")

	query := `
		WITH latest_unique_reg AS (
			SELECT DISTINCT ON (registration) icao, registration, 
			operator, type, icao_type, "group", 
			category, tag1, tag2, tag3, image_link_1, 
			image_link_2, image_link_3,
					hex, flight, seen, seen_epoch
			FROM interesting_aircraft_seen
			WHERE "group" = $1
			ORDER BY registration, seen DESC
		)
		SELECT *
		FROM latest_unique_reg
		ORDER BY seen DESC
		LIMIT $2`

	rows, err := s.pg.db.Query(context.Background(), query, group, limit)
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

func (s *APIServer) getFastestAircraft(c *gin.Context) {
	limit := s.getLimit("record_holder_table_limit")

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
	limit := s.getLimit("record_holder_table_limit")

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
	limit := s.getLimit("record_holder_table_limit")

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
	limit := s.getLimit("record_holder_table_limit")

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

func (s *APIServer) getTopAircraftTypes(c *gin.Context, period string, flightoraircraft string) {
	var query string
	var timeFilter string
	var innerFilter string
	var innerQuery string

	switch period {
	case "year":
		timeFilter = `age(now(), first_seen) <= INTERVAL '1 year' AND`
	case "month":
		timeFilter = `age(now(), first_seen) <= INTERVAL '1 month' AND`
	case "day":
		timeFilter = `age(now(), first_seen) <= INTERVAL '1 day' AND`
	default:
		timeFilter = ""
	}
	innerFilter = `WHERE ` + timeFilter + ` t IS NOT NULL AND t != ''`

	switch flightoraircraft {
	case "aircraft":
		innerQuery = `(SELECT t, hex FROM aircraft_data ` + innerFilter + `GROUP BY t, hex)`
	case "flights":
		innerQuery = `aircraft_data ` + innerFilter
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid flightoraircraft parameter. Use 'flights' or 'aircraft'"})
		return
	}

	query = `SELECT
					t,
					count,
					ROUND(count * 100.0 / SUM(count) OVER(), 0) as percentage
				FROM (
					SELECT t, Count(t) as count
					FROM ` + innerQuery + `
					GROUP BY t ORDER BY count DESC
				) top_15
				ORDER BY count DESC LIMIT 15`

	rows, err := s.pg.db.Query(context.Background(), query)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	defer rows.Close()

	aircraft := []gin.H{}

	for rows.Next() {
		var aircraft_type string
		var count int
		var percentage float64

		err := rows.Scan(&aircraft_type, &count, &percentage)

		if err != nil {
			continue
		}

		aircraft = append(aircraft, gin.H{
			"aircraft_type": aircraft_type,
			"count":         count,
			"percentage":    percentage,
		})
	}

	c.JSON(http.StatusOK, aircraft)

}

func (s *APIServer) getTopRoutes(c *gin.Context) {
	limit := s.getLimit("route_table_limit")

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
	limit := s.getLimit("route_table_limit")

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
	limit := s.getLimit("route_table_limit")

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
	limit := s.getLimit("route_table_limit")

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
	limit := s.getLimit("route_table_limit")

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
	limit := s.getLimit("route_table_limit")

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

func (s *APIServer) getChartFlightsOverTime(c *gin.Context, period string) {
	tz := s.getTimezone(c)
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.UTC
	}

	var query string
	var seriesID, label, periodUnit string

	switch period {
	case "year":
		seriesID = "flights_year"
		label = "Flights Past Year"
		periodUnit = "month"
		query = `WITH months AS (
				SELECT generate_series(
					DATE_TRUNC('month', CURRENT_DATE - INTERVAL '12 months'),
					DATE_TRUNC('month', CURRENT_DATE),
					'1 month'
				)::date AS month
				),
				counts AS (
				SELECT
					DATE_TRUNC('month', first_seen)::date AS month,
					COUNT(*) AS count
				FROM aircraft_data
				WHERE first_seen >= DATE_TRUNC('month', CURRENT_DATE - INTERVAL '12 months')
					AND first_seen < DATE_TRUNC('month', CURRENT_DATE) + INTERVAL '1 month'
				GROUP BY 1
				)
				SELECT
				m.month::timestamptz,
				COALESCE(c.count, 0) AS count
				FROM months m
				LEFT JOIN counts c USING (month)
				ORDER BY m.month;`
	case "month":
		seriesID = "flights_month"
		label = "Flights Past Month"
		periodUnit = "day"
		query = `WITH days AS (
				SELECT generate_series(
					CURRENT_DATE - INTERVAL '1 month',
					CURRENT_DATE,
					'1 day'
				)::date AS day
				),
				counts AS (
				SELECT
					DATE(first_seen) AS day,
					COUNT(*) AS count
				FROM aircraft_data
				WHERE first_seen >= CURRENT_DATE - INTERVAL '1 month'
					AND first_seen < CURRENT_DATE + INTERVAL '1 day'
				GROUP BY 1
				)
				SELECT
					d.day::timestamptz,
					COALESCE(c.count, 0) AS count
				FROM days d
				LEFT JOIN counts c USING (day)
				ORDER BY d.day;`
	case "day":
		seriesID = "flights_day"
		label = "Flights Past 24 Hours"
		periodUnit = "hour"
		query = `WITH end_hour AS (
				SELECT date_trunc('hour', CURRENT_TIMESTAMP AT TIME ZONE 'UTC') AS h,
				       CURRENT_TIMESTAMP AT TIME ZONE 'UTC' AS now
				)
				SELECT
				gs AS hour,
				COALESCE(c.count, 0) AS count
				FROM generate_series(
					(SELECT h FROM end_hour) - interval '23 hours',
					(SELECT h FROM end_hour),
					interval '1 hour'
					) AS gs
				LEFT JOIN (
				SELECT date_trunc('hour', first_seen) AS hour, COUNT(*) AS count
				FROM aircraft_data, end_hour
				WHERE first_seen >= (SELECT h FROM end_hour) - interval '23 hours'
					AND first_seen <= (SELECT now FROM end_hour)
				GROUP BY 1
				) c ON c.hour = gs
				ORDER BY gs;`
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid period. Use 'year', 'month', or 'day'"})
		return
	}

	var rows pgx.Rows
	ctx := context.Background()

	if period == "month" || period == "year" {
		tx, err := s.pg.db.Begin(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer tx.Rollback(ctx)

		_, err = tx.Exec(ctx, fmt.Sprintf("SET LOCAL TIME ZONE '%s'", tz))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		rows, err = tx.Query(ctx, query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		defer tx.Commit(ctx)

	} else {
		var err error
		rows, err = s.pg.db.Query(ctx, query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	defer rows.Close()

	results := []ChartPoint{}

	for rows.Next() {
		var timeVal time.Time
		var count int

		err := rows.Scan(&timeVal, &count)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		results = append(results, ChartPoint{
			X: timeVal.In(loc),
			Y: float64(count),
		})
	}

	c.JSON(http.StatusOK, ChartResponse{
		Series: []ChartSeries{
			{
				ID:     seriesID,
				Label:  label,
				Unit:   "count",
				Points: results,
			},
		},
		X: ChartXAxisMeta{
			Type: "time",
			Unit: periodUnit,
		},
		Meta: ChartMeta{
			GeneratedAt: time.Now(),
		},
	})
}

func (s *APIServer) getChartAircraftOverTime(c *gin.Context, period string) {
	tz := s.getTimezone(c)
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.UTC
	}

	var query string
	var seriesID, label, periodUnit string

	switch period {
	case "year":
		seriesID = "aircraft_year"
		label = "Aircraft Past Year"
		periodUnit = "month"
		query = `WITH months AS (
				SELECT generate_series(
					DATE_TRUNC('month', CURRENT_DATE - INTERVAL '12 months'),
					DATE_TRUNC('month', CURRENT_DATE),
					'1 month'
				)::date AS month
				),
				counts AS (
				SELECT
					DATE_TRUNC('month', first_seen)::date AS month,
					COUNT(DISTINCT hex) AS count
				FROM aircraft_data
				WHERE first_seen >= DATE_TRUNC('month', CURRENT_DATE - INTERVAL '12 months')
					AND first_seen <  DATE_TRUNC('month', CURRENT_DATE) + INTERVAL '1 month'
				GROUP BY 1
				)
				SELECT
				m.month::timestamptz,
				COALESCE(c.count, 0) AS count
				FROM months m
				LEFT JOIN counts c USING (month)
				ORDER BY m.month;`
	case "month":
		seriesID = "aircraft_month"
		label = "Aircraft Past Month"
		periodUnit = "day"
		query = `WITH days AS (
				SELECT generate_series(
					CURRENT_DATE - INTERVAL '1 month',
					CURRENT_DATE,
					'1 day'
				)::date AS day
				),
				counts AS (
				SELECT
					DATE(first_seen) AS day,
					COUNT(DISTINCT hex) AS count
				FROM aircraft_data
				WHERE first_seen >= CURRENT_DATE - INTERVAL '1 month'
					AND first_seen < CURRENT_DATE + INTERVAL '1 day'
				GROUP BY 1
				)
				SELECT
					d.day::timestamptz,
				COALESCE(c.count, 0) AS count
				FROM days d
				LEFT JOIN counts c USING (day)
				ORDER BY d.day;`
	case "day":
		seriesID = "aircraft_day"
		label = "Aircraft Past 24 Hours"
		periodUnit = "hour"
		query = `WITH end_hour AS (
				SELECT
					date_trunc('hour', CURRENT_TIMESTAMP) AS h,
					CURRENT_TIMESTAMP AS now
				)
				SELECT
				gs AS hour,
				COALESCE(c.count, 0) AS count
				FROM generate_series(
				(SELECT h FROM end_hour) - interval '23 hours',
				(SELECT h FROM end_hour),
				interval '1 hour'
				) AS gs
				LEFT JOIN (
				SELECT
					date_trunc('hour', first_seen) AS hour,
					COUNT(DISTINCT hex) AS count
				FROM aircraft_data, end_hour
				WHERE first_seen >= (SELECT h FROM end_hour) - interval '23 hours'
					AND first_seen <= (SELECT now FROM end_hour)
				GROUP BY 1
				) c ON c.hour = gs
				ORDER BY gs;`
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid period. Use 'year', 'month', or 'day'"})
		return
	}

	var rows pgx.Rows
	ctx := context.Background()

	if period == "month" || period == "year" {
		tx, err := s.pg.db.Begin(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer tx.Rollback(ctx)

		_, err = tx.Exec(ctx, fmt.Sprintf("SET LOCAL TIME ZONE '%s'", tz))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		rows, err = tx.Query(ctx, query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		defer tx.Commit(ctx)

	} else {
		var err error
		rows, err = s.pg.db.Query(ctx, query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	defer rows.Close()

	results := []ChartPoint{}

	for rows.Next() {
		var timeVal time.Time
		var count int

		err := rows.Scan(&timeVal, &count)
		if err != nil {
			continue
		}

		results = append(results, ChartPoint{
			X: timeVal.In(loc),
			Y: float64(count),
		})
	}

	c.JSON(http.StatusOK, ChartResponse{
		Series: []ChartSeries{
			{
				ID:     seriesID,
				Label:  label,
				Unit:   "count",
				Points: results,
			},
		},
		X: ChartXAxisMeta{
			Type: "time",
			Unit: periodUnit,
		},
		Meta: ChartMeta{
			GeneratedAt: time.Now(),
		},
	})
}

func (s *APIServer) getVersion(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version": version,
		"commit":  commit,
		"date":    date,
	})
}

func (s *APIServer) getLimit(settingKey ...string) int {

	if len(settingKey) == 1 && settingKey[0] != "" {
		setting, err := s.settings.GetSetting(settingKey[0])
		if err == nil {
			limit, err := strconv.Atoi(setting.SettingValue)
			if err == nil {
				return limit
			}
		}
	}

	// Default if no setting
	return 5

}

func (s *APIServer) getCountry() string {
	return os.Getenv("DOMESTIC_COUNTRY_ISO")
}

func (s *APIServer) getSettings(c *gin.Context) {
	settings, err := s.settings.GetAllSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}

func (s *APIServer) updateSettings(c *gin.Context) {
	var updates map[string]string
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := s.settings.UpdateSettings(updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return updated settings
	settings, err := s.settings.GetAllSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}
