package cases

import (
	"context"

	en "github.com/100bench/subscription_aggregator/internal/entities"
	"github.com/pkg/errors"
)

type ServiceProvider struct {
	storage SubRepository
}

func NewServiceProvider(storage SubRepository) (*ServiceProvider, error) {
	if storage == nil {
		return nil, errors.Wrap(en.ErrNilDependency, "storage")
	}
	return &ServiceProvider{storage}, nil
}

func (s *ServiceProvider) CreateSubscription(ctx context.Context, subscription en.Subscription) error {
	err := s.storage.CreateSub(ctx, subscription)
	if err != nil {
		return errors.Wrap(err, "storage.CreateSub")
	}
	return nil
}

func (s *ServiceProvider) GetSubscription(ctx context.Context, userID string, serviceName string) (en.Subscription, error) {
	sub, err := s.storage.GetSub(ctx, userID, serviceName)
	if err != nil {
		return en.Subscription{}, errors.Wrap(err, "storage.GetSub")
	}
	return sub, nil
}

func (s *ServiceProvider) UpdateSubscription(ctx context.Context, userID string, serviceName string, price *int, startDate *string, endDate *string) error {
	err := s.storage.UpdateSub(ctx, userID, serviceName, price, startDate, endDate)
	if err != nil {
		return errors.Wrap(err, "storage.UpdateSub")
	}
	return nil
}

func (s *ServiceProvider) DeleteSubscription(ctx context.Context, userID string, serviceName string) error {
	err := s.storage.DeleteSub(ctx, userID, serviceName)
	if err != nil {
		return errors.Wrap(err, "storage.DeleteSub")
	}
	return nil
}

func (s *ServiceProvider) GetListSubscriptions(ctx context.Context, userID string) ([]en.Subscription, error) {
	subs, err := s.storage.GetListSubs(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, "storage.GetListSubs")
	}
	return subs, nil
}

func (s *ServiceProvider) GetTotalCostByPeriod(ctx context.Context, userID string, serviceName string, startDate string, endDate string) (int, error) {
	cost, err := s.storage.GetTotalByPeriod(ctx, userID, serviceName, startDate, endDate)
	if err != nil {
		return 0, errors.Wrap(err, "storage.GetTotalCostByPeriod")
	}
	return cost, nil
}
