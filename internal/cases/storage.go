package cases

import (
	"context"

	en "github.com/100bench/subscription_aggregator/internal/entities"
)

type SubRepository interface {
	CreateSub(ctx context.Context, subscription en.Subscription) (int64, error)
	GetSubByID(ctx context.Context, id int64) (en.Subscription, error)
	UpdateSubByID(ctx context.Context, id int64, price *int, startDate *string, endDate *string) error
	DeleteSubByID(ctx context.Context, id int64) error
	GetListSubs(ctx context.Context, userID, serviceName string, page, pageSize int) ([]en.Subscription, int, error)
	GetTotalByPeriod(ctx context.Context, userID string, serviceName string, startDate, endDate string) (int, error)
}
