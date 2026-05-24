package main

import (
	"os"
	"strconv"
	"time"
)

const (
	defaultIngestIntervalSec = 2
	defaultMetricsPollMs     = 2000
	minIngestIntervalSec     = 1
	maxIngestIntervalSec     = 60
	minMetricsPollMs         = 1000
	maxMetricsPollMs         = 60000
)

func getIngestInterval() time.Duration {
	sec := defaultIngestIntervalSec
	if raw := os.Getenv("SKYSTATS_INGEST_INTERVAL_SEC"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			sec = parsed
		}
	}
	if sec < minIngestIntervalSec {
		sec = minIngestIntervalSec
	}
	if sec > maxIngestIntervalSec {
		sec = maxIngestIntervalSec
	}
	return time.Duration(sec) * time.Second
}

func getMetricsPollMs() int {
	ms := defaultMetricsPollMs
	if raw := os.Getenv("SKYSTATS_METRICS_POLL_MS"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			ms = parsed
		}
	}
	if ms < minMetricsPollMs {
		ms = minMetricsPollMs
	}
	if ms > maxMetricsPollMs {
		ms = maxMetricsPollMs
	}
	return ms
}
