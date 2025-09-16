package cases

import (
	"context"
	en "github.com/100bench/subscription_aggregator/internal/entities"
	"github.com/pkg/errors"
)

type ServiceProvider struct {
	provider SubProvider
	storage  SubscriptionRepository
}

func NewServiceProvider(provider SubProvider, storage SubscriptionRepository) (*ServiceProvider, error) {
	if provider == nil {
		return nil, errors.Wrap(en.ErrNilDependency, "nil dependency: provider")
	}
	return &ServiceProvider{provider: provider, storage: storage}, nil
}

func (s *ServiceProvider) GetSubscription(ctx context.Context, userID string) error {
	subscription, err := s.provider.FetchSub(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "usecase ServiceProvider.provider.FetchSub")
	}
	err = s.storage.CreateSub(ctx, subscription)
	if err != nil {
		return errors.Wrap(err, "usecase ServiceProvider.storage.CreateSub")
	}
	return nil
}
