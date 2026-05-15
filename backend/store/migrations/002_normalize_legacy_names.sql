-- +goose Up
-- Normalize legacy country names that pre-date the iso_mapping.go rename so
-- that loaded DB rows match the canonical autocomplete names. Without this,
-- a user with a legacy `Cape Verde` or `Palestine` row could add the new
-- canonical `Cabo Verde` / `Palestine State` from the autocomplete and the
-- frontend duplicate check (raw string compare) would miss it, inflating
-- state.visited.length.
--
-- UPDATE OR IGNORE silently skips rows where the target (user_id, new name)
-- already exists; the follow-up DELETE then removes the now-redundant legacy
-- row so each user has exactly one canonical entry.
UPDATE OR IGNORE visits SET country_name = 'Cabo Verde' WHERE country_name = 'Cape Verde';
DELETE FROM visits WHERE country_name = 'Cape Verde';

UPDATE OR IGNORE visits SET country_name = 'Palestine State' WHERE country_name = 'Palestine';
DELETE FROM visits WHERE country_name = 'Palestine';

-- +goose Down
-- No reversible Down — the original legacy values are lost. Restore from a
-- DB backup if rollback is required.
SELECT 1;
