package public

import (
	"context"

	en "github.com/100bench/subscription_aggregator/internal/entities"
)

type PublicService interface {
	CreateSubscription(ctx context.Context, subscription en.Subscription) (int64, error)
	GetSubscriptionByID(ctx context.Context, id int64) (en.Subscription, error)
	UpdateSubscriptionByID(ctx context.Context, id int64, price *int, startDate *string, endDate *string) error
	DeleteSubscriptionByID(ctx context.Context, id int64) error
	GetListSubscriptions(ctx context.Context, userID, serviceName string, page, pageSize int) ([]en.Subscription, int, error)
	GetTotalCostByPeriod(ctx context.Context, userID string, serviceName string, startDate string, endDate string) (int, error)
}
