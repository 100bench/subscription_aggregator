-- Add CHECK constraint to ensure end_date is after or equal to start_date
ALTER TABLE subscriptions
ADD CONSTRAINT check_date_range CHECK (end_date IS NULL OR end_date >= start_date);
