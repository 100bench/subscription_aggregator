package postgres

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"

	en "github.com/100bench/subscription_aggregator/internal/entities"
)

type PgxStorage struct {
	pool *pgxpool.Pool
}

func NewPgxClient(ctx context.Context, dsn string) (*PgxStorage, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, errors.Wrap(err, "pgxpool.ParseConfig")
	}
	cfg.MaxConns = 10
	cfg.MinConns = 2
	cfg.MaxConnLifetime = time.Hour
	cfg.MaxConnIdleTime = time.Minute

	pool, err := pgxpool.ConnectConfig(ctx, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "pgxpool.ConnectConfig")
	}
	return &PgxStorage{pool: pool}, nil
}

func (p *PgxStorage) Close() {
	p.pool.Close()
}

func (p *PgxStorage) CreateSub(ctx context.Context, sub en.Subscription) error {
	log.Printf("INFO: CreateSub for user %s, service %s", sub.UserID, sub.ServiceName)
	const q = `
		INSERT INTO subscriptions (user_id, service_name, price, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := p.pool.Exec(ctx, q, sub.UserID, sub.ServiceName, sub.Price, sub.StartDate, sub.EndDate)
	if err != nil {
		log.Printf("ERROR: failed to create subscription for user %s, service %s: %v", sub.UserID, sub.ServiceName, err)
		return errors.Wrap(err, "PgxStorage.CreateSub")
	}
	log.Printf("INFO: Subscription created for user %s, service %s", sub.UserID, sub.ServiceName)
	return nil
}

func (p *PgxStorage) GetSub(ctx context.Context, userID, serviceName string) (en.Subscription, error) {
	log.Printf("INFO: GetSub for user %s, service %s", userID, serviceName)
	const q = `
		SELECT user_id, service_name, price, start_date, end_date FROM subscriptions
		WHERE user_id = $1 AND service_name = $2
	`
	var sub en.Subscription
	err := p.pool.QueryRow(ctx, q, userID, serviceName).Scan(
		&sub.UserID,
		&sub.ServiceName,
		&sub.Price,
		&sub.StartDate,
		&sub.EndDate,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("WARN: Subscription not found for user %s, service %s: %v", userID, serviceName, err)
			return en.Subscription{}, errors.Wrap(en.ErrSubscriptionNotFound, "PgxStorage.GetSub")
		}
		log.Printf("ERROR: failed to get subscription for user %s, service %s: %v", userID, serviceName, err)
		return en.Subscription{}, errors.Wrap(err, "PgxStorage.GetSub")
	}
	log.Printf("INFO: Subscription retrieved for user %s, service %s", userID, serviceName)
	return sub, nil
}

func (p *PgxStorage) GetListSubs(ctx context.Context, userId string) ([]en.Subscription, error) {
	log.Printf("INFO: GetListSubs for user %s", userId)
	const q = `
		SELECT user_id, service_name, price, start_date, end_date FROM subscriptions
		WHERE user_id = $1
	`
	rows, err := p.pool.Query(ctx, q, userId)
	if err != nil {
		log.Printf("ERROR: failed to get list of subscriptions for user %s: %v", userId, err)
		return nil, errors.Wrap(err, "PgxStorage.GetListSubs")
	}
	defer rows.Close()

	var subscriptions []en.Subscription
	for rows.Next() {
		var sub en.Subscription
		if err := rows.Scan(
			&sub.UserID,
			&sub.ServiceName,
			&sub.Price,
			&sub.StartDate,
			&sub.EndDate,
		); err != nil {
			log.Printf("ERROR: failed to scan subscription row for user %s: %v", userId, err)
			return nil, errors.Wrap(err, "PgxStorage.GetListSubs.Scan")
		}
		subscriptions = append(subscriptions, sub)
	}
	if rows.Err() != nil {
		log.Printf("ERROR: rows iteration error for user %s: %v", userId, rows.Err())
		return nil, errors.Wrap(rows.Err(), "PgxStorage.GetListSubs.RowsError")
	}
	log.Printf("INFO: Retrieved %d subscriptions for user %s", len(subscriptions), userId)
	return subscriptions, nil
}

func (p *PgxStorage) UpdateSub(ctx context.Context, userID, serviceName string, price *int, startDate *string, endDate *string) error {
	log.Printf("INFO: UpdateSub for user %s, service %s", userID, serviceName)
	const q = `
        UPDATE subscriptions
        SET price = COALESCE($3, price),
            start_date = COALESCE($4, start_date),
            end_date = COALESCE($5, end_date)
        WHERE user_id = $1 AND service_name = $2
    `
	commandTag, err := p.pool.Exec(ctx, q, userID, serviceName, price, startDate, endDate)
	if err != nil {
		log.Printf("ERROR: failed to update subscription for user %s, service %s: %v", userID, serviceName, err)
		return errors.Wrap(err, "PgxStorage.UpdateSub")
	}
	if commandTag.RowsAffected() == 0 {
		log.Printf("WARN: Subscription not found for update for user %s, service %s", userID, serviceName)
		return errors.Wrap(en.ErrSubscriptionNotFound, "PgxStorage.UpdateSub")
	}
	log.Printf("INFO: Subscription updated for user %s, service %s", userID, serviceName)
	return nil
}

func (p *PgxStorage) DeleteSub(ctx context.Context, userID, serviceName string) error {
	log.Printf("INFO: DeleteSub for user %s, service %s", userID, serviceName)
	const q = `
		DELETE FROM subscriptions
		WHERE user_id = $1 AND service_name = $2
	`
	commandTag, err := p.pool.Exec(ctx, q, userID, serviceName)
	if err != nil {
		log.Printf("ERROR: failed to delete subscription for user %s, service %s: %v", userID, serviceName, err)
		return errors.Wrap(err, "PgxStorage.DeleteSub")
	}
	if commandTag.RowsAffected() == 0 {
		log.Printf("WARN: Subscription not found for delete for user %s, service %s", userID, serviceName)
		return errors.Wrap(en.ErrSubscriptionNotFound, "PgxStorage.DeleteSub")
	}
	log.Printf("INFO: Subscription deleted for user %s, service %s", userID, serviceName)
	return nil
}

func (p *PgxStorage) GetTotalByPeriod(ctx context.Context, userID string, serviceName string, startDateStr, endDateStr string) (int, error) {
	log.Printf("INFO: GetTotalByPeriod input userID=%s service=%q start=%s end=%s", userID, serviceName, startDateStr, endDateStr)

	startDate, err := time.Parse("01-2006", startDateStr)
	if err != nil {
		log.Printf("ERROR: parse start_date=%s (want MM-YYYY): %v", startDateStr, err)
		return 0, fmt.Errorf("invalid start date format (MM-YYYY): %w", err)
	}
	endDate, err := time.Parse("01-2006", endDateStr)
	if err != nil {
		log.Printf("ERROR: parse end_date=%s (want MM-YYYY): %v", endDateStr, err)
		return 0, fmt.Errorf("invalid end date format (MM-YYYY): %w", err)
	}

	// Нормализуем конец периода на конец месяца для наглядности логов
	endDateLastDay := endDate.AddDate(0, 1, -1)
	log.Printf("DEBUG: normalized period start=%s end=%s(end-of-month=%s)", startDate.Format(time.RFC3339), endDate.Format(time.RFC3339), endDateLastDay.Format(time.RFC3339))

	var (
		q    string
		args []interface{}
	)

	if serviceName != "" {
		q = `
			SELECT COALESCE(SUM(price), 0) AS total
			FROM subscriptions
			WHERE user_id = $1
			  AND service_name = $2
			  AND to_date(start_date, 'MM-YYYY') >= $3::date
			  AND to_date(start_date, 'MM-YYYY') <= $4::date
		`
		args = []interface{}{userID, serviceName, startDate, endDate}
	} else {
		q = `
			SELECT COALESCE(SUM(price), 0) AS total
			FROM subscriptions
			WHERE user_id = $1
			  AND to_date(start_date, 'MM-YYYY') >= $2::date
			  AND to_date(start_date, 'MM-YYYY') <= $3::date
		`
		args = []interface{}{userID, startDate, endDate}
	}

	log.Printf("DEBUG: SQL=%q args=%v", q, args)

	var total int
	if err := p.pool.QueryRow(ctx, q, args...).Scan(&total); err != nil {
		log.Printf("ERROR: QueryRow failed userID=%s service=%q: %v", userID, serviceName, err)
		return 0, errors.Wrap(err, "PgxStorage.GetTotalByPeriod")
	}

	log.Printf("INFO: GetTotalByPeriod result userID=%s service=%q total=%d", userID, serviceName, total)
	return total, nil
}
