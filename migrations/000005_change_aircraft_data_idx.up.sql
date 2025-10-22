-- Lower the threshold for when autovacuum runs, but increase the sleep time
ALTER TABLE aircraft_data SET (
  autovacuum_vacuum_scale_factor = 0.1,
  autovacuum_vacuum_cost_delay = 10
);

-- Drop previous bloated / rarely used index
DROP INDEX IF EXISTS idx_aircraft_data_hex_last_seen;

-- Drop unused index
DROP INDEX IF EXISTS aircraft_data_hex;