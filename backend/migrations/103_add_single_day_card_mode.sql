-- Add single_day_card_mode for subscription groups.
-- When enabled, each redeem resets quota usage and validity from redeem time.

ALTER TABLE groups
    ADD COLUMN IF NOT EXISTS single_day_card_mode BOOLEAN NOT NULL DEFAULT FALSE;

COMMENT ON COLUMN groups.single_day_card_mode IS
    'Single-day card mode: each redeem resets quota windows and overwrites validity from redeem time';
