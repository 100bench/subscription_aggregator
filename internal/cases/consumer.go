package cases

import (
	"context"
	en "github.com/100bench/subscription_aggregator/internal/entities"
)

type SubConsumer interface {
	SendSub(ctx context.Context, subscription en.Subscription) error
}
