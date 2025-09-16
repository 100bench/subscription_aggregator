package cases

import (
	"context"
	en "github.com/100bench/subscription_aggregator/internal/entities"
	"github.com/pkg/errors"
)

type Consumer struct {
	consumer SubConsumer
	storage  SubscriptionRepository
}

func NewConsumer(consumer SubConsumer, storage SubscriptionRepository) (*Consumer, error) {
	if consumer == nil {
		return nil, errors.Wrap(en.ErrNilDependency, "nil dependency: consumer")
	}
	return &Consumer{consumer: consumer, storage: storage}, nil
}

func (c *Consumer) SendSubscription(ctx context.Context, userID string) error {
	subscription, err := c.storage.GetSub(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "usecase Consumer.storage.GetSub")
	}
	err = c.consumer.SendSub(ctx, subscription)
	if err != nil {
		return errors.Wrap(err, "usecase Consumer.consumer.SendSub")
	}
	return nil
}

func (c *Consumer) SendListSubscriptions(ctx context.Context, userID string) error {
	subscriptions, err := c.storage.GetListSubs(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "usecase Consumer.storage.GetListSubs")
	}
	for _, subscription := range subscriptions {
		err = c.consumer.SendSub(ctx, subscription)
		if err != nil {
			return errors.Wrap(err, "usecase Consumer.consumer.SendSub")
		}
	}
	return nil
}
