CREATE TABLE subscriptions(
    id bigserial PRIMARY KEY,
    user_id uuid NOT NULL,
    service_name text NOT NULL,
    price integer NOT NULL,
    start_date text NOT NULL,
    end_date text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (user_id, service_name, start_date)
);

CREATE INDEX idx_subs_user ON subscriptions(user_id);
CREATE INDEX idx_subs_user_service ON subscriptions(user_id, service_name);