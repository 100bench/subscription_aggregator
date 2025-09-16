package cases

import (
	"context"

	pkg "github.com/100bench/subscription_aggregator/pkg/dto"
)

// SubscriptionService определяет контракт для операций с подписками на уровне бизнес-логики.
type SubscriptionService interface {
	CreateSubscription(ctx context.Context, req pkg.CreateSubRequest) (pkg.SubscriptionDTO, error)
	GetSubscription(ctx context.Context, userID, serviceName string) (pkg.SubscriptionDTO, error)
	GetAllSubscriptions(ctx context.Context, userID string) (pkg.GetSubsResponse, error)
	UpdateSubscription(ctx context.Context, userID, serviceName string, req pkg.UpdateSubRequest) (pkg.SubscriptionDTO, error)
	DeleteSubscription(ctx context.Context, userID, serviceName string) error
	GetTotalSubscriptionCost(ctx context.Context, req pkg.GetTotalCostRequest) (pkg.GetTotalCostResponse, error)
}
