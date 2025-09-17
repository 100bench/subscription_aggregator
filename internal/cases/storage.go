package cases

import (
	"context"

	en "github.com/100bench/subscription_aggregator/internal/entities"
)

type SubRepository interface {
	CreateSub(ctx context.Context, subscription en.Subscription) error
	GetSub(ctx context.Context, userID, serviceName string) (en.Subscription, error)
	UpdateSub(ctx context.Context, userID, serviceName string, price *int, startDate *string, endDate *string) error
	DeleteSub(ctx context.Context, userID, serviceName string) error
	GetListSubs(ctx context.Context, userId string) ([]en.Subscription, error)
	GetTotalByPeriod(ctx context.Context, userID string, serviceName string, startDate, endDate string) (int, error)
}
