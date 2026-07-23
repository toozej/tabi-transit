# Mobile public Mapbox token handling

This guidance concerns a native-app public token only. It does not authorize
reading, printing, committing, or otherwise inspecting local `.env` files.

Use a distinct `pk...` token for each environment and grant only the public
scopes required by the selected native map capability. A public token can be
embedded in an app, but it remains account-linked usage data: never place it in
fixtures, tests, source code, screenshots, or issue text. Do not use a secret
`sk...` token or a server/download token in the mobile app.

Place the development token in the ignored mobile local environment file using
the documented public Expo configuration variable. Confirm that the file is
ignored with `git check-ignore` rather than printing its contents. Use separate
staging and production tokens, monitor account usage, and revoke/rotate a token
if it is disclosed outside its intended distribution.

Before enabling Mapbox Search, Geocoding, offline content, telemetry, or
persistent place storage, close D-004 with the selected product's terms,
attribution, storage, telemetry, scope, and budget evidence. This repository
does not assert those external approvals.
