package cases

import (
	"context"
	en "github.com/100bench/subscription_aggregator/internal/entities"
)

type SubProvider interface {
	FetchSub(ctx context.Context, userID string) (en.Subscription, error)
}
