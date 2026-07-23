# ADR-0011: Feature-gated notification delivery with encrypted token boundary

- Status: Accepted
- Date: 2026-07-23
- Owners: Architecture, Security, Privacy, Mobile
- Evidence required: D-017 Expo project/credential and isolated receipt evidence; D-018 retention, consent, and subscription-policy approval

## Context

Notifications introduce installation credentials, platform push tokens, user
preferences, time-sensitive delivery, and deletion obligations. This
workstation cannot exercise a device permission flow or a real Expo Push
delivery, and no project or credential status is inferred.

## Decision

Use installation-owned, opaque credentials; store only a verifier for the
credential and authenticated ciphertext plus a hash for a push token. Keep the
API contract, mobile settings, persistence model, and worker abstraction in
place, but feature-disable routes, token registration, and delivery until D-017
and D-018 are cleared. The worker must be a no-op when disabled, use unique
delivery keys and leased claims when enabled, honor quiet hours and expiry, and
never retry an expired time-sensitive notification.

## Consequences

The mobile app can present an accessible, contextual settings flow without
asking OS permission at launch. There is no real token, request, receipt, or
delivery claim. Backend logs and notification payloads exclude credentials,
tokens, coordinates, and full itineraries. The worker binary is built into the
runtime image but is not enabled in Compose until its runtime composition and
external evidence are approved.

## Rollback / forward fix

Leave the notifications feature disabled to remove the runtime path. Future
enablement must use an additive configuration and Compose profile, preserve the
schema's deletion cascade, and include a receipt/invalid-token test before
turning on delivery.

## Validation

OpenAPI lint/generation, Go unit/race tests for validation, encryption and
worker policies, migration tests, and mobile typecheck/Vitest pass. Device
permission, real push, receipt, and deletion-on-provider evidence are not yet
run.
