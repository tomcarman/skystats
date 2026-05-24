package main

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
)

type CachedStatsService struct {
	pg *postgres
}

func NewCachedStatsService(pg *postgres) *CachedStatsService {
	return &CachedStatsService{pg: pg}
}

func (s *CachedStatsService) updateCachedStat(key string, value int) error {
	_, err := s.pg.db.Exec(context.Background(),
		`INSERT INTO cached_stats (stat_key, stat_value, last_updated)
		VALUES ($1, $2, NOW())
		ON CONFLICT (stat_key)
		DO UPDATE SET stat_value = $2, last_updated = NOW()`,
		key, value)
	return err
}

func (s *CachedStatsService) GetCachedTotalAircraft() (int, error) {
	var value int
	var lastUpdated time.Time
	err := s.pg.db.QueryRow(context.Background(),
		`SELECT stat_value, last_updated FROM cached_stats
		WHERE stat_key = 'total_aircraft'`).Scan(&value, &lastUpdated)

	if err == nil && time.Since(lastUpdated) < 24*time.Hour {
		log.Debug().
			Int("total_aircraft", value).
			Dur("cache_age", time.Since(lastUpdated)).
			Msg("Returning cached total_aircraft")
		return value, nil
	}

	log.Info().Msg("Cache stale or missing, recalculating total_aircraft...")

	var totalAircraft int
	err = s.pg.db.QueryRow(context.Background(),
		"SELECT COUNT(DISTINCT hex) FROM aircraft_data").Scan(&totalAircraft)
	if err != nil {
		log.Error().Err(err).Msg("Failed to count total aircraft")
		return 0, err
	}

	// Update cache
	err = s.updateCachedStat("total_aircraft", totalAircraft)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update cached total_aircraft")
		return totalAircraft, nil
	}

	log.Info().
		Int("total_aircraft", totalAircraft).
		Msg("Successfully recalculated and cached total_aircraft")

	return totalAircraft, nil
}

func (s *CachedStatsService) GetCachedTotalFlights() (int, error) {
	var value int
	var lastUpdated time.Time
	err := s.pg.db.QueryRow(context.Background(),
		`SELECT stat_value, last_updated FROM cached_stats
		WHERE stat_key = 'total_flights'`).Scan(&value, &lastUpdated)

	if err == nil && time.Since(lastUpdated) < 24*time.Hour {
		log.Debug().
			Int("total_flights", value).
			Dur("cache_age", time.Since(lastUpdated)).
			Msg("Returning cached total_flights")
		return value, nil
	}

	log.Info().Msg("Cache stale or missing, recalculating total_flights...")

	var totalFlights int
	err = s.pg.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM aircraft_data").Scan(&totalFlights)
	if err != nil {
		log.Error().Err(err).Msg("Failed to count total flights")
		return 0, err
	}

	if err := s.updateCachedStat("total_flights", totalFlights); err != nil {
		log.Error().Err(err).Msg("Failed to update cached total_flights")
		return totalFlights, nil
	}

	log.Info().
		Int("total_flights", totalFlights).
		Msg("Successfully recalculated and cached total_flights")

	return totalFlights, nil
}

// RefreshCachedTotals recomputes expensive all-time counts (used by the stats ticker).
func (s *CachedStatsService) RefreshCachedTotals() {
	if _, err := s.GetCachedTotalFlights(); err != nil {
		log.Error().Err(err).Msg("Failed to refresh cached total_flights")
	}
	if _, err := s.GetCachedTotalAircraft(); err != nil {
		log.Error().Err(err).Msg("Failed to refresh cached total_aircraft")
	}
}
