package main

import (
	"context"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

type APIServer struct {
	pg   *postgres
	port string
}

func NewAPIServer(pg *postgres) *APIServer {
	port := os.Getenv("API_PORT")
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

	// CORS middleware
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
			stats.GET("/fastest", s.handleFastestAircraft)
			stats.GET("/slowest", s.handleSlowestAircraft)
			stats.GET("/highest", s.handleHighestAircraft)
			stats.GET("/lowest", s.handleLowestAircraft)
			stats.GET("/interesting", s.handleInterestingAircraft)
			stats.GET("/general", s.handleGeneralStats)
		}
	}

	// Serve static files (must come after API routes)
	r.Static("/static", "../web")
	r.StaticFile("/", "../web/index.html")

	r.Run("0.0.0.0:" + s.port)
}

func (s *APIServer) handleFastestAircraft(c *gin.Context) {
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

	var aircraft []gin.H
	for rows.Next() {
		var hex, flight, registration, aircraftType string
		var firstSeen, lastSeen interface{}
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

func (s *APIServer) handleSlowestAircraft(c *gin.Context) {
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

	var aircraft []gin.H
	for rows.Next() {
		var hex, flight, registration, aircraftType string
		var firstSeen, lastSeen interface{}
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

func (s *APIServer) handleHighestAircraft(c *gin.Context) {
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

	var aircraft []gin.H
	for rows.Next() {
		var hex, flight, registration, aircraftType string
		var firstSeen, lastSeen interface{}
		var barometricAltitude, geometricAltitude int

		err := rows.Scan(&hex, &flight, &registration, &aircraftType, &firstSeen, 
			&lastSeen, &barometricAltitude, &geometricAltitude)
		if err != nil {
			continue
		}

		aircraft = append(aircraft, gin.H{
			"hex":                  hex,
			"flight":               flight,
			"registration":         registration,
			"type":                 aircraftType,
			"first_seen":           firstSeen,
			"last_seen":            lastSeen,
			"barometric_altitude":  barometricAltitude,
			"geometric_altitude":   geometricAltitude,
		})
	}

	c.JSON(http.StatusOK, aircraft)
}

func (s *APIServer) handleLowestAircraft(c *gin.Context) {
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

	var aircraft []gin.H
	for rows.Next() {
		var hex, flight, registration, aircraftType string
		var firstSeen, lastSeen interface{}
		var barometricAltitude, geometricAltitude int

		err := rows.Scan(&hex, &flight, &registration, &aircraftType, &firstSeen, 
			&lastSeen, &barometricAltitude, &geometricAltitude)
		if err != nil {
			continue
		}

		aircraft = append(aircraft, gin.H{
			"hex":                  hex,
			"flight":               flight,
			"registration":         registration,
			"type":                 aircraftType,
			"first_seen":           firstSeen,
			"last_seen":            lastSeen,
			"barometric_altitude":  barometricAltitude,
			"geometric_altitude":   geometricAltitude,
		})
	}

	c.JSON(http.StatusOK, aircraft)
}

func (s *APIServer) handleInterestingAircraft(c *gin.Context) {
	limit := s.getLimit(c)
	
	query := `
		SELECT icao, registration, operator, type, icao_type, "group", 
			   category, hex, flight, seen, seen_epoch
		FROM interesting_aircraft_seen 
		ORDER BY seen DESC 
		LIMIT $1`

	rows, err := s.pg.db.Query(context.Background(), query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var aircraft []gin.H
	for rows.Next() {
		var icao, registration, operator, aircraftType, icaoType, group, category string
		var hex, flight string
		var seen interface{}
		var seenEpoch float64

		err := rows.Scan(&icao, &registration, &operator, &aircraftType, &icaoType, 
			&group, &category, &hex, &flight, &seen, &seenEpoch)
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
			"hex":          hex,
			"flight":       flight,
			"seen":         seen,
			"seen_epoch":   seenEpoch,
		})
	}

	c.JSON(http.StatusOK, aircraft)
}

func (s *APIServer) handleGeneralStats(c *gin.Context) {
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

	// Fastest recorded speed
	var fastestSpeed float64
	err = s.pg.db.QueryRow(context.Background(), 
		"SELECT MAX(ground_speed) FROM fastest_aircraft").Scan(&fastestSpeed)
	if err == nil {
		stats["fastest_speed"] = fastestSpeed
	}

	// Highest altitude
	var highestAltitude int
	err = s.pg.db.QueryRow(context.Background(), 
		"SELECT MAX(barometric_altitude) FROM highest_aircraft").Scan(&highestAltitude)
	if err == nil {
		stats["highest_altitude"] = highestAltitude
	}

	c.JSON(http.StatusOK, stats)
}

func (s *APIServer) getLimit(c *gin.Context) int {
	limitStr := c.DefaultQuery("limit", "10")
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		return 10
	}
	
	if limit > 100 {
		return 100 // max limit
	}
	
	return limit
}