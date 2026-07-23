-- Isolated proof only: this is not a production migration.
\o /dev/null
CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TEMP TABLE vehicle_fixture (
  id text PRIMARY KEY,
  mode text NOT NULL,
  point geography(Point, 4326) NOT NULL
);

INSERT INTO vehicle_fixture (id, mode, point) VALUES
  ('bus-nearest', 'bus', ST_SetSRID(ST_MakePoint(-122.6700, 45.5200), 4326)::geography),
  ('bus-second',  'bus', ST_SetSRID(ST_MakePoint(-122.6680, 45.5200), 4326)::geography),
  ('rail-nearest','light_rail', ST_SetSRID(ST_MakePoint(-122.6700, 45.5210), 4326)::geography),
  ('rail-second', 'light_rail', ST_SetSRID(ST_MakePoint(-122.6660, 45.5200), 4326)::geography);

\o
WITH input AS (
  SELECT ST_SetSRID(ST_MakePoint(-122.6700, 45.5200), 4326)::geography AS point
), nearby AS (
  SELECT v.id, v.mode, round(ST_Distance(v.point, input.point))::integer AS distance_meters,
         row_number() OVER (PARTITION BY v.mode ORDER BY v.point <-> input.point, v.id) AS mode_rank
  FROM vehicle_fixture v CROSS JOIN input
  WHERE ST_DWithin(v.point, input.point, 500)
)
SELECT id, mode, distance_meters, mode_rank
FROM nearby WHERE mode_rank <= 1
ORDER BY mode, id;
