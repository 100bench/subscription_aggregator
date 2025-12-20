-- Remove CHECK constraint for date range
ALTER TABLE subscriptions
DROP CONSTRAINT IF EXISTS check_date_range;
