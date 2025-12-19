package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"

	en "github.com/100bench/subscription_aggregator/internal/entities"
	"github.com/100bench/subscription_aggregator/internal/pkg/billing"
	"github.com/100bench/subscription_aggregator/internal/pkg/logger"
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

func (p *PgxStorage) CreateSub(ctx context.Context, sub en.Subscription) (int64, error) {
	logger.Log.Info("CreateSub", "user_id", sub.UserID, "service", sub.ServiceName)

	// Parse dates from MM-YYYY to time.Time
	startDate, err := time.Parse("01-2006", sub.StartDate)
	if err != nil {
		logger.Log.Error("Invalid start_date format", "start_date", sub.StartDate, "error", err)
		return 0, fmt.Errorf("invalid start_date format (want MM-YYYY): %w", err)
	}

	var endDate *time.Time
	if sub.EndDate != "" {
		parsed, err := time.Parse("01-2006", sub.EndDate)
		if err != nil {
			logger.Log.Error("Invalid end_date format", "end_date", sub.EndDate, "error", err)
			return 0, fmt.Errorf("invalid end_date format (want MM-YYYY): %w", err)
		}
		endDate = &parsed
	}

	const q = `
		INSERT INTO subscriptions (user_id, service_name, price, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	var id int64
	err = p.pool.QueryRow(ctx, q, sub.UserID, sub.ServiceName, sub.Price, startDate, endDate).Scan(&id)
	if err != nil {
		logger.Log.Error("Failed to create subscription", "user_id", sub.UserID, "service", sub.ServiceName, "error", err)
		return 0, errors.Wrap(err, "PgxStorage.CreateSub")
	}
	logger.Log.Info("Subscription created", "user_id", sub.UserID, "service", sub.ServiceName, "id", id)
	return id, nil
}

func (p *PgxStorage) GetSubByID(ctx context.Context, id int64) (en.Subscription, error) {
	logger.Log.Info("GetSubByID", "id", id)
	const q = `
		SELECT id, user_id, service_name, price, start_date, end_date FROM subscriptions
		WHERE id = $1
	`
	var sub en.Subscription
	var startDate time.Time
	var endDateNullable *time.Time

	err := p.pool.QueryRow(ctx, q, id).Scan(
		&sub.ID,
		&sub.UserID,
		&sub.ServiceName,
		&sub.Price,
		&startDate,
		&endDateNullable,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Log.Warn("Subscription not found", "id", id)
			return en.Subscription{}, errors.Wrap(en.ErrSubscriptionNotFound, "PgxStorage.GetSubByID")
		}
		logger.Log.Error("Failed to get subscription", "id", id, "error", err)
		return en.Subscription{}, errors.Wrap(err, "PgxStorage.GetSubByID")
	}

	// Convert dates to MM-YYYY format
	sub.StartDate = startDate.Format("01-2006")
	if endDateNullable != nil {
		sub.EndDate = endDateNullable.Format("01-2006")
	} else {
		sub.EndDate = ""
	}

	logger.Log.Info("Subscription retrieved", "id", id)
	return sub, nil
}

func (p *PgxStorage) GetListSubs(ctx context.Context, userID, serviceName string, page, pageSize int) ([]en.Subscription, int, error) {
	logger.Log.Info("GetListSubs", "user_id", userID, "service_name", serviceName, "page", page, "page_size", pageSize)

	// Validate and normalize pagination parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// Build WHERE clause based on filters
	var countQ, q string
	var countArgs, queryArgs []interface{}

	if serviceName != "" {
		countQ = `SELECT COUNT(*) FROM subscriptions WHERE user_id = $1 AND service_name = $2`
		countArgs = []interface{}{userID, serviceName}

		q = `
			SELECT id, user_id, service_name, price, start_date, end_date FROM subscriptions
			WHERE user_id = $1 AND service_name = $2
			ORDER BY created_at DESC
			LIMIT $3 OFFSET $4
		`
	} else {
		countQ = `SELECT COUNT(*) FROM subscriptions WHERE user_id = $1`
		countArgs = []interface{}{userID}

		q = `
			SELECT id, user_id, service_name, price, start_date, end_date FROM subscriptions
			WHERE user_id = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`
	}

	// Get total count
	var total int
	err := p.pool.QueryRow(ctx, countQ, countArgs...).Scan(&total)
	if err != nil {
		logger.Log.Error("Failed to count subscriptions", "user_id", userID, "error", err)
		return nil, 0, errors.Wrap(err, "PgxStorage.GetListSubs.Count")
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Build query args
	if serviceName != "" {
		queryArgs = []interface{}{userID, serviceName, pageSize, offset}
	} else {
		queryArgs = []interface{}{userID, pageSize, offset}
	}

	// Get paginated subscriptions
	rows, err := p.pool.Query(ctx, q, queryArgs...)
	if err != nil {
		logger.Log.Error("Failed to get list of subscriptions", "user_id", userID, "error", err)
		return nil, 0, errors.Wrap(err, "PgxStorage.GetListSubs")
	}
	defer rows.Close()

	var subscriptions []en.Subscription
	for rows.Next() {
		var sub en.Subscription
		var startDate time.Time
		var endDateNullable *time.Time

		if err := rows.Scan(
			&sub.ID,
			&sub.UserID,
			&sub.ServiceName,
			&sub.Price,
			&startDate,
			&endDateNullable,
		); err != nil {
			logger.Log.Error("Failed to scan subscription row", "user_id", userID, "error", err)
			return nil, 0, errors.Wrap(err, "PgxStorage.GetListSubs.Scan")
		}

		// Convert dates to MM-YYYY format
		sub.StartDate = startDate.Format("01-2006")
		if endDateNullable != nil {
			sub.EndDate = endDateNullable.Format("01-2006")
		} else {
			sub.EndDate = ""
		}

		subscriptions = append(subscriptions, sub)
	}
	if rows.Err() != nil {
		logger.Log.Error("Rows iteration error", "user_id", userID, "error", rows.Err())
		return nil, 0, errors.Wrap(rows.Err(), "PgxStorage.GetListSubs.RowsError")
	}
	logger.Log.Info("Retrieved subscriptions", "count", len(subscriptions), "total", total, "user_id", userID)
	return subscriptions, total, nil
}

func (p *PgxStorage) UpdateSubByID(ctx context.Context, id int64, price *int, startDate *string, endDate *string) error {
	logger.Log.Info("UpdateSubByID", "id", id)

	// Convert string dates to time.Time if provided
	var startDateTime *time.Time
	if startDate != nil {
		parsed, err := time.Parse("01-2006", *startDate)
		if err != nil {
			logger.Log.Error("Invalid start_date format", "start_date", *startDate, "error", err)
			return fmt.Errorf("invalid start_date format: %w", err)
		}
		startDateTime = &parsed
	}

	var endDateTime *time.Time
	if endDate != nil && *endDate != "" {
		parsed, err := time.Parse("01-2006", *endDate)
		if err != nil {
			logger.Log.Error("Invalid end_date format", "end_date", *endDate, "error", err)
			return fmt.Errorf("invalid end_date format: %w", err)
		}
		endDateTime = &parsed
	}

	const q = `
        UPDATE subscriptions
        SET price = COALESCE($2, price),
            start_date = COALESCE($3, start_date),
            end_date = COALESCE($4, end_date)
        WHERE id = $1
    `
	commandTag, err := p.pool.Exec(ctx, q, id, price, startDateTime, endDateTime)
	if err != nil {
		logger.Log.Error("Failed to update subscription", "id", id, "error", err)
		return errors.Wrap(err, "PgxStorage.UpdateSubByID")
	}
	if commandTag.RowsAffected() == 0 {
		logger.Log.Warn("Subscription not found for update", "id", id)
		return errors.Wrap(en.ErrSubscriptionNotFound, "PgxStorage.UpdateSubByID")
	}
	logger.Log.Info("Subscription updated", "id", id)
	return nil
}

func (p *PgxStorage) DeleteSubByID(ctx context.Context, id int64) error {
	logger.Log.Info("DeleteSubByID", "id", id)
	const q = `
		DELETE FROM subscriptions
		WHERE id = $1
	`
	commandTag, err := p.pool.Exec(ctx, q, id)
	if err != nil {
		logger.Log.Error("Failed to delete subscription", "id", id, "error", err)
		return errors.Wrap(err, "PgxStorage.DeleteSubByID")
	}
	if commandTag.RowsAffected() == 0 {
		logger.Log.Warn("Subscription not found for delete", "id", id)
		return errors.Wrap(en.ErrSubscriptionNotFound, "PgxStorage.DeleteSubByID")
	}
	logger.Log.Info("Subscription deleted", "id", id)
	return nil
}

func (p *PgxStorage) GetTotalByPeriod(ctx context.Context, userID string, serviceName string, startDateStr, endDateStr string) (int, error) {
	logger.Log.Info("GetTotalByPeriod", "user_id", userID, "service", serviceName, "start", startDateStr, "end", endDateStr)

	// Parse query period dates
	queryStart, err := time.Parse("01-2006", startDateStr)
	if err != nil {
		logger.Log.Error("Failed to parse start_date", "start_date", startDateStr, "error", err)
		return 0, fmt.Errorf("invalid start date format (MM-YYYY): %w", err)
	}
	queryEnd, err := time.Parse("01-2006", endDateStr)
	if err != nil {
		logger.Log.Error("Failed to parse end_date", "end_date", endDateStr, "error", err)
		return 0, fmt.Errorf("invalid end date format (MM-YYYY): %w", err)
	}

	// Build query to fetch all relevant subscriptions
	var (
		q    string
		args []interface{}
	)

	if serviceName != "" {
		q = `
			SELECT service_name, price, start_date, end_date
			FROM subscriptions
			WHERE user_id = $1 AND service_name = $2
		`
		args = []interface{}{userID, serviceName}
	} else {
		q = `
			SELECT service_name, price, start_date, end_date
			FROM subscriptions
			WHERE user_id = $1
		`
		args = []interface{}{userID}
	}

	rows, err := p.pool.Query(ctx, q, args...)
	if err != nil {
		logger.Log.Error("Failed to query subscriptions", "user_id", userID, "service", serviceName, "error", err)
		return 0, errors.Wrap(err, "PgxStorage.GetTotalByPeriod.Query")
	}
	defer rows.Close()

	totalCost := 0

	for rows.Next() {
		var (
			svcName          string
			price            int
			subStart         time.Time
			subEndNullable   *time.Time
		)

		if err := rows.Scan(&svcName, &price, &subStart, &subEndNullable); err != nil {
			logger.Log.Error("Failed to scan subscription row", "error", err)
			return 0, errors.Wrap(err, "PgxStorage.GetTotalByPeriod.Scan")
		}

		// Determine subscription end date (use a far future date if null)
		subEnd := time.Date(9999, 12, 31, 0, 0, 0, 0, time.UTC)
		if subEndNullable != nil {
			subEnd = *subEndNullable
		}

		// Calculate overlapping months
		months := billing.CalculateMonths(subStart, subEnd, queryStart, queryEnd)

		if months > 0 {
			cost := price * months
			totalCost += cost
			logger.Log.Debug("Subscription overlap",
				"service", svcName,
				"price", price,
				"sub_start", subStart.Format("01-2006"),
				"sub_end", subEnd.Format("01-2006"),
				"months", months,
				"cost", cost)
		}
	}

	if rows.Err() != nil {
		logger.Log.Error("Rows iteration error", "error", rows.Err())
		return 0, errors.Wrap(rows.Err(), "PgxStorage.GetTotalByPeriod.RowsError")
	}

	logger.Log.Info("GetTotalByPeriod result", "user_id", userID, "service", serviceName, "total", totalCost)
	return totalCost, nil
}
