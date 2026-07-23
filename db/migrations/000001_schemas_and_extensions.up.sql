BEGIN;

CREATE EXTENSION IF NOT EXISTS postgis;

CREATE SCHEMA IF NOT EXISTS catalog;
CREATE SCHEMA IF NOT EXISTS transit;
CREATE SCHEMA IF NOT EXISTS realtime;
CREATE SCHEMA IF NOT EXISTS ops;

CREATE TYPE transit.mode AS ENUM (
  'bus', 'light_rail', 'streetcar', 'commuter_rail', 'ferry', 'other', 'unknown'
);
CREATE TYPE catalog.feed_version_status AS ENUM ('staged', 'active', 'superseded', 'failed');
CREATE TYPE realtime.freshness_status AS ENUM ('fresh', 'aging', 'stale', 'unknown');

COMMIT;
