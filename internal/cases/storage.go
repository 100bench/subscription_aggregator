package cases

import (
	"context"

	en "github.com/100bench/subscription_aggregator/internal/entities"
)

type SubscriptionRepository interface {
	CreateSub(ctx context.Context, subscription en.Subscription) error
	GetSub(ctx context.Context, userID string) (en.Subscription, error)
	UpdateSub(ctx context.Context, subscription en.Subscription) error
	DeleteSub(ctx context.Context, userID string) error
	GetListSubs(ctx context.Context, userId string) ([]en.Subscription, error)
	GetTotalCostByPeriod(ctx context.Context, userID string, serviceName string, startDate, endDate string) (int, error)
}
