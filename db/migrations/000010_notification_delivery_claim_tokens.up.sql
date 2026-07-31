BEGIN;

-- A claim token distinguishes successive leases of the same delivery. Every
-- worker transition must present the token returned by Claim, so a worker that
-- outlives its lease cannot overwrite a later worker's result.
ALTER TABLE app.notification_deliveries
  ADD COLUMN claim_token uuid;

COMMIT;
