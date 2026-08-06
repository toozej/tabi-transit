# Web release artifact

`apps/web` produces immutable, content-hashed static files in `dist/`. The
release process copies one verified build into this directory (or sets
`TABI_WEB_ASSET_DIR` to its versioned equivalent) before Caddy is reloaded.

The rollback procedure is to repoint `TABI_WEB_ASSET_DIR` at the prior verified
artifact and reload Caddy. API deployments retain their compatibility window;
a web rollback must not require an API-contract rollback.

Do not place environment files, provider keys, source maps intended to remain
private, or any rider data in this directory.
