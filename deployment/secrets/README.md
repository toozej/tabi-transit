# Secret-file convention

Production secrets live outside this repository in `/etc/tabi/secrets`. The directory is root-owned and mode `0700`; each file is root-owned and mode `0400` or `0600`. Set `TABI_SECRETS_DIR=/etc/tabi/secrets` in `/opt/tabi/.env`.

Compose mounts these files read-only and applications receive only their `_FILE` paths:

- `postgres_password` → `DATABASE_PASSWORD_FILE` / `POSTGRES_PASSWORD_FILE`
- `trimet_app_id` → `TRIMET_APP_ID_FILE`
- `mapbox_server_token` → `MAPBOX_SERVER_TOKEN_FILE`
- `installation_auth_key` → `INSTALLATION_AUTH_KEY_FILE`
- `restic_password` and `restic_environment` are host-only inputs for backup jobs.

Do not create real secret files, copy secret values into `.env`, image layers, logs, or CI output. Local validation creates temporary placeholder files outside the repository.
