# Production secret files

Create these files under `/etc/tabi/secrets`, owned by root and mode `0400` or `0600`:

- `postgres_password`
- `trimet_app_id`
- `mapbox_server_token`
- `installation_auth_key`
- `restic_password`
- `restic_environment`

Do not put secret values in this repository. The Go configuration loader must support `_FILE` settings. The restic environment file contains repository/backend credentials and is sourced only by the backup service.
