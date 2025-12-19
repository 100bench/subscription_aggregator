-- Add index for pagination queries (GetListSubs with ORDER BY created_at DESC)
CREATE INDEX IF NOT EXISTS idx_subs_user_created_at ON subscriptions(user_id, created_at DESC);

-- Add index for date range queries in GetTotalByPeriod
CREATE INDEX IF NOT EXISTS idx_subs_dates ON subscriptions(user_id, start_date, end_date);
