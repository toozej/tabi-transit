#!/usr/bin/env bash
set -euo pipefail

root_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
migrations_dir="$root_dir/db/migrations"
database_url=${TABI_DATABASE_URL:-${DATABASE_URL:-}}

if [[ -z "$database_url" ]]; then
  printf '%s\n' 'Set TABI_DATABASE_URL (or DATABASE_URL) to the target PostgreSQL database.' >&2
  exit 2
fi
command -v psql >/dev/null || { printf '%s\n' 'psql is required to run migrations.' >&2; exit 127; }

psql "$database_url" -v ON_ERROR_STOP=1 -c '
CREATE TABLE IF NOT EXISTS public.tabi_schema_migrations (
  filename text PRIMARY KEY,
  applied_at timestamptz NOT NULL DEFAULT now()
);'
psql "$database_url" -v ON_ERROR_STOP=1 -c '
DO $$
BEGIN
  IF to_regclass('"'"'public.schema_migrations'"'"') IS NOT NULL THEN
    INSERT INTO public.tabi_schema_migrations (filename)
    SELECT version FROM public.schema_migrations
    ON CONFLICT (filename) DO NOTHING;
  END IF;
END $$;'
for migration in "$migrations_dir"/*.up.sql; do
  version=$(basename "$migration")
  applied=$(psql "$database_url" -At -v ON_ERROR_STOP=1 -c "SELECT 1 FROM public.tabi_schema_migrations WHERE filename = '$version'")
  if [[ "$applied" == "1" ]]; then
    continue
  fi
  printf 'Applying %s\n' "$version"
  psql "$database_url" -v ON_ERROR_STOP=1 -f "$migration"
  psql "$database_url" -v ON_ERROR_STOP=1 -c "INSERT INTO public.tabi_schema_migrations (filename) VALUES ('$version')"
done
