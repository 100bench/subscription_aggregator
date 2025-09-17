package public

import (
	"context"

	en "github.com/100bench/subscription_aggregator/internal/entities"
)

type PublicService interface {
	CreateSubscription(ctx context.Context, subscription en.Subscription) error
	GetSubscription(ctx context.Context, userID string, serviceName string) (en.Subscription, error)
	UpdateSubscription(ctx context.Context, userID string, serviceName string, price *int, startDate *string, endDate *string) error
	DeleteSubscription(ctx context.Context, userID string, serviceName string) error
	GetListSubscriptions(ctx context.Context, userID string) ([]en.Subscription, error)
	GetTotalCostByPeriod(ctx context.Context, userID string, serviceName string, startDate string, endDate string) (int, error)
}
