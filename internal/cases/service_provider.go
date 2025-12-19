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

func (s *ServiceProvider) CreateSubscription(ctx context.Context, subscription en.Subscription) (int64, error) {
	id, err := s.storage.CreateSub(ctx, subscription)
	if err != nil {
		return 0, errors.Wrap(err, "storage.CreateSub")
	}
	return id, nil
}

func (s *ServiceProvider) GetSubscriptionByID(ctx context.Context, id int64) (en.Subscription, error) {
	sub, err := s.storage.GetSubByID(ctx, id)
	if err != nil {
		return en.Subscription{}, errors.Wrap(err, "storage.GetSubByID")
	}
	return sub, nil
}

func (s *ServiceProvider) UpdateSubscriptionByID(ctx context.Context, id int64, price *int, startDate *string, endDate *string) error {
	err := s.storage.UpdateSubByID(ctx, id, price, startDate, endDate)
	if err != nil {
		return errors.Wrap(err, "storage.UpdateSubByID")
	}
	return nil
}

func (s *ServiceProvider) DeleteSubscriptionByID(ctx context.Context, id int64) error {
	err := s.storage.DeleteSubByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, "storage.DeleteSubByID")
	}
	return nil
}

func (s *ServiceProvider) GetListSubscriptions(ctx context.Context, userID, serviceName string, page, pageSize int) ([]en.Subscription, int, error) {
	subs, total, err := s.storage.GetListSubs(ctx, userID, serviceName, page, pageSize)
	if err != nil {
		return nil, 0, errors.Wrap(err, "storage.GetListSubs")
	}
	return subs, total, nil
}

func (s *ServiceProvider) GetTotalCostByPeriod(ctx context.Context, userID string, serviceName string, startDate string, endDate string) (int, error) {
	cost, err := s.storage.GetTotalByPeriod(ctx, userID, serviceName, startDate, endDate)
	if err != nil {
		return 0, errors.Wrap(err, "storage.GetTotalCostByPeriod")
	}
	return cost, nil
}
