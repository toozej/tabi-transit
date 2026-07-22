---
doc_id: TAB-PLAN-019
title: "Authoritative Reference Sources"
status: implementation-ready
last_updated: 2026-07-22
intended_agents: ["all-agents"]
depends_on: ["TAB-PLAN-000"]
---


# Reference Sources

Accessed 2026-07-22. Re-check versions, terms, pricing, and URLs before implementation.

## Expo

- https://docs.expo.dev/
- https://docs.expo.dev/develop/development-builds/introduction/
- https://docs.expo.dev/workflow/continuous-native-generation/
- https://docs.expo.dev/guides/new-architecture/
- https://docs.expo.dev/guides/permissions/
- https://docs.expo.dev/router/introduction/
- https://docs.expo.dev/versions/latest/sdk/sqlite/
- https://docs.expo.dev/versions/latest/sdk/location/
- https://docs.expo.dev/versions/latest/sdk/background-task/
- https://docs.expo.dev/push-notifications/push-notifications-setup/
- https://docs.expo.dev/build/introduction/
- https://docs.expo.dev/deploy/submit-to-app-stores/

## Mapbox and RNMapbox

- https://github.com/rnmapbox/maps
- https://rnmapbox.github.io/
- https://rnmapbox.github.io/docs/install
- https://rnmapbox.github.io/docs/components/MapView
- https://rnmapbox.github.io/docs/components/MarkerView
- https://rnmapbox.github.io/docs/components/ShapeSource
- https://rnmapbox.github.io/docs/components/SymbolLayer
- https://docs.mapbox.com/api/search/geocoding/
- https://docs.mapbox.com/api/search/search-box/
- https://docs.mapbox.com/api/navigation/directions/
- https://docs.mapbox.com/accounts/guides/tokens/
- https://docs.mapbox.com/help/getting-started/attribution/
- https://www.mapbox.com/pricing/

## TriMet/GTFS

- https://developer.trimet.org/
- https://developer.trimet.org/ws_docs/
- https://developer.trimet.org/GTFS.shtml
- https://developer.trimet.org/gtfs_ext.shtml
- https://developer.trimet.org/gis/
- https://gtfs.org/documentation/overview/
- https://gtfs.org/documentation/schedule/reference/
- https://gtfs.org/documentation/realtime/reference/
- https://gtfs.org/documentation/realtime/realtime-best-practices/
- https://gtfs.org/documentation/realtime/language-bindings/golang/
- https://github.com/MobilityData/gtfs-realtime-bindings
- https://github.com/MobilityData/gtfs-validator

## Optional sources

- https://www.rosecitytransit.org/
- https://www.rosecitytransit.org/tools/
- https://www.rosecitytransit.org/systemmapper/
- https://www.rosecitytransit.org/transitmapper/
- https://portlandstreetcar.org/
- https://retro.umoiq.com/googleMap/googleMap.jsp?a=portland-sc

These pages do not by themselves establish a public API or redistribution license.

## Backend, Docker Compose, backups and testing

- https://go.dev/doc/
- https://www.postgresql.org/docs/
- https://postgis.net/docs/
- https://postgis.net/docs/ST_DWithin.html
- https://docs.opentripplanner.org/
- https://opentelemetry.io/docs/
- https://prometheus.io/docs/
- https://docs.docker.com/engine/install/
- https://docs.docker.com/compose/
- https://docs.docker.com/compose/how-tos/production/
- https://docs.docker.com/engine/logging/configure/
- https://caddyserver.com/docs/quick-starts/reverse-proxy
- https://caddyserver.com/docs/running
- https://restic.readthedocs.io/en/stable/
- https://restic.readthedocs.io/en/stable/030_preparing_a_new_repo.html
- https://vitest.dev/guide/
- https://callstack.github.io/react-native-testing-library/
- https://docs.maestro.dev/
- https://docs.github.com/actions
- https://docs.github.com/actions/tutorials/publish-packages/publish-docker-images
- https://docs.github.com/actions/how-tos/deploy/configure-and-manage-deployments/control-deployments

## Optional Fly.io path

- https://fly.io/docs/about/cost-management/
- https://fly.io/docs/about/pricing/
- https://fly.io/docs/languages-and-frameworks/dockerfile/
- https://fly.io/docs/reference/configuration/
- https://fly.io/docs/launch/processes/
- https://fly.io/docs/apps/secrets/
- https://fly.io/docs/volumes/overview/
- https://fly.io/docs/postgres/getting-started/what-you-should-know/
- https://fly.io/docs/machines/guides-examples/multi-container-machines/

Fly.io currently documents no standing free account/free tier. Treat the platform as an optional trial or low-cost target and verify pricing before each deployment.

## Re-check policy

Verify current stable compatibility, prefer primary docs, review terms/pricing, pin versions in ADR/tool files, and update this list when sources move.
