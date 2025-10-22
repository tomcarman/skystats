-- Recreate the old indexes that were dropped
CREATE INDEX IF NOT EXISTS idx_aircraft_data_hex_last_seen
ON aircraft_data (hex, last_seen_epoch DESC);

CREATE INDEX IF NOT EXISTS aircraft_data_hex
ON aircraft_data (hex);

-- Reset autovacuum settings to defaults
ALTER TABLE aircraft_data RESET (
  autovacuum_vacuum_scale_factor,
  autovacuum_vacuum_cost_delay
);
