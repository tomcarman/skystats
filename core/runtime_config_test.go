package main

import (
	"os"
	"testing"
	"time"
)

func TestGetIngestInterval(t *testing.T) {
	os.Unsetenv("SKYSTATS_INGEST_INTERVAL_SEC")
	if got := getIngestInterval(); got != 2*time.Second {
		t.Fatalf("got %v, want 2s", got)
	}

	t.Setenv("SKYSTATS_INGEST_INTERVAL_SEC", "5")
	if got := getIngestInterval(); got != 5*time.Second {
		t.Fatalf("got %v, want 5s", got)
	}

	t.Setenv("SKYSTATS_INGEST_INTERVAL_SEC", "0")
	if got := getIngestInterval(); got != time.Second {
		t.Fatalf("got %v, want 1s minimum", got)
	}
}

func TestGetMetricsPollMs(t *testing.T) {
	os.Unsetenv("SKYSTATS_METRICS_POLL_MS")
	if got := getMetricsPollMs(); got != 2000 {
		t.Fatalf("got %d, want 2000", got)
	}

	t.Setenv("SKYSTATS_METRICS_POLL_MS", "5000")
	if got := getMetricsPollMs(); got != 5000 {
		t.Fatalf("got %d, want 5000", got)
	}
}
