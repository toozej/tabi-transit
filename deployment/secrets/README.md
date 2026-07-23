# Secret-file convention

Production secrets live outside this repository in `/etc/tabi/secrets`. The directory is root-owned and mode `0700`; each file is root-owned and mode `0400` or `0600`. Set `TABI_SECRETS_DIR=/etc/tabi/secrets` in `/opt/tabi/.env`.

Compose mounts these files read-only and applications receive only their `_FILE` paths:

- `postgres_password` → `DATABASE_PASSWORD_FILE` / `POSTGRES_PASSWORD_FILE`
- `database_url` → `TABI_DATABASE_URL_FILE` (backend and migration jobs)
- `trimet_app_id` → `TRIMET_APP_ID_FILE`
- `restic_password` and `restic_environment` are host-only inputs for backup jobs.

`database_url` is a PostgreSQL connection URL for the private Compose network,
for example `postgres://tabi:PASSWORD@postgres:5432/tabi?sslmode=disable`.
Generate it from the same value as `postgres_password`; do not put either value
in `.env`. Mapbox server access and notification authentication are not mounted
until their corresponding backend capabilities exist.

Do not create real secret files, copy secret values into `.env`, image layers, logs, or CI output. Local validation creates temporary placeholder files outside the repository.
