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

func (p *PgxStorage) UpdateSub(ctx context.Context, sub en.Subscription) error {
	log.Printf("INFO: UpdateSub for user %s, service %s", sub.UserID, sub.ServiceName)
	const q = `
		UPDATE subscriptions
		SET price = $3, start_date = $4, end_date = $5
		WHERE user_id = $1 AND service_name = $2
	`
	commandTag, err := p.pool.Exec(ctx, q, sub.UserID, sub.ServiceName, sub.Price, sub.StartDate, sub.EndDate)
	if err != nil {
		log.Printf("ERROR: failed to update subscription for user %s, service %s: %v", sub.UserID, sub.ServiceName, err)
		return errors.Wrap(err, "PgxStorage.UpdateSub")
	}
	if commandTag.RowsAffected() == 0 {
		log.Printf("WARN: Subscription not found for update for user %s, service %s", sub.UserID, sub.ServiceName)
		return errors.Wrap(en.ErrSubscriptionNotFound, "PgxStorage.UpdateSub")
	}
	log.Printf("INFO: Subscription updated for user %s, service %s", sub.UserID, sub.ServiceName)
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

func (p *PgxStorage) GetTotalCostByPeriod(ctx context.Context, userID string, serviceName string, startDateStr, endDateStr string) (int, error) {
	log.Printf("INFO: GetTotalCostByPeriod requested for user %s, service %s, from %s to %s", userID, serviceName, startDateStr, endDateStr)

	startDate, err := time.Parse("2006-01", startDateStr)
	if err != nil {
		log.Printf("ERROR: invalid start date format %s: %v", startDateStr, err)
		return 0, fmt.Errorf("invalid start date format: %w", err)
	}

	endDate, err := time.Parse("2006-01", endDateStr)
	if err != nil {
		log.Printf("ERROR: invalid end date format %s: %v", endDateStr, err)
		return 0, fmt.Errorf("invalid end date format: %w", err)
	}
	endDate = endDate.AddDate(0, 1, -1)

	var q string
	var args []interface{}

	if serviceName != "" {
		q = `
			SELECT COALESCE(SUM(price), 0) FROM subscriptions
			WHERE user_id = $1 AND service_name = $2 AND start_date >= $3 AND start_date <= $4
		`
		args = []interface{}{userID, serviceName, startDate, endDate}
	} else {
		q = `
			SELECT COALESCE(SUM(price), 0) FROM subscriptions
			WHERE user_id = $1 AND start_date >= $2 AND start_date <= $3
		`
		args = []interface{}{userID, startDate, endDate}
	}

	var totalCost int
	err = p.pool.QueryRow(ctx, q, args...).Scan(&totalCost)
	if err != nil {
		log.Printf("ERROR: failed to get total cost from database for user %s: %v", userID, err)
		return 0, errors.Wrap(err, "PgxStorage.GetTotalCostByPeriod")
	}

	log.Printf("INFO: Total cost for user %s is %d", userID, totalCost)
	return totalCost, nil
}
