package cases

import (
	"context"
	"fmt"
	"log"

	en "github.com/100bench/subscription_aggregator/internal/entities"
	pkg "github.com/100bench/subscription_aggregator/pkg/dto"
)

type serviceProvider struct {
	repo SubscriptionRepository
}

func NewSubscriptionService(repo SubscriptionRepository) SubscriptionService {
	return &serviceProvider{
		repo: repo,
	}
}

func (s *serviceProvider) CreateSubscription(ctx context.Context, req pkg.CreateSubRequest) (pkg.SubscriptionDTO, error) {
	log.Printf("INFO: CreateSubscription requested for user %s, service %s", req.UserId, req.ServiceName)
	sub := en.Subscription{
		UserID:      req.UserId,
		ServiceName: req.ServiceName,
		Price:       req.Price,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
	}

	if err := s.repo.CreateSub(ctx, sub); err != nil {
		log.Printf("ERROR: failed to create subscription for user %s, service %s: %v", req.UserId, req.ServiceName, err)
		return pkg.SubscriptionDTO{}, fmt.Errorf("failed to create subscription: %w", err)
	}

	log.Printf("INFO: Subscription created successfully for user %s, service %s", req.UserId, req.ServiceName)
	return pkg.SubscriptionDTO{
		UserId:      sub.UserID,
		ServiceName: sub.ServiceName,
		Price:       sub.Price,
		StartDate:   sub.StartDate,
		EndDate:     sub.EndDate,
	}, nil
}

func (s *serviceProvider) GetSubscription(ctx context.Context, userID, serviceName string) (pkg.SubscriptionDTO, error) {
	log.Printf("INFO: GetSubscription requested for user %s, service %s", userID, serviceName)
	sub, err := s.repo.GetSub(ctx, userID)
	if err != nil {
		log.Printf("ERROR: failed to get subscription for user %s, service %s: %v", userID, serviceName, err)
		return pkg.SubscriptionDTO{}, fmt.Errorf("failed to get subscription: %w", err)
	}

	if sub.ServiceName != serviceName {
		log.Printf("WARN: Service name mismatch for user %s. Expected %s, got %s", userID, serviceName, sub.ServiceName)
		return pkg.SubscriptionDTO{}, en.ErrSubscriptionNotFound
	}

	log.Printf("INFO: Subscription retrieved for user %s, service %s", userID, serviceName)
	return pkg.SubscriptionDTO{
		UserId:      sub.UserID,
		ServiceName: sub.ServiceName,
		Price:       sub.Price,
		StartDate:   sub.StartDate,
		EndDate:     sub.EndDate,
	}, nil
}

func (s *serviceProvider) GetAllSubscriptions(ctx context.Context, userID string) (pkg.GetSubsResponse, error) {
	log.Printf("INFO: GetAllSubscriptions requested for user %s", userID)
	subs, err := s.repo.GetListSubs(ctx, userID)
	if err != nil {
		log.Printf("ERROR: failed to get all subscriptions for user %s: %v", userID, err)
		return pkg.GetSubsResponse{}, fmt.Errorf("failed to get all subscriptions: %w", err)
	}

	var dtos []pkg.SubscriptionDTO
	for _, sub := range subs {
		dtos = append(dtos, pkg.SubscriptionDTO{
			UserId:      sub.UserID,
			ServiceName: sub.ServiceName,
			Price:       sub.Price,
			StartDate:   sub.StartDate,
			EndDate:     sub.EndDate,
		})
	}

	log.Printf("INFO: Retrieved %d subscriptions for user %s", len(dtos), userID)
	return pkg.GetSubsResponse{Subscriptions: dtos}, nil
}

func (s *serviceProvider) UpdateSubscription(ctx context.Context, userID, serviceName string, req pkg.UpdateSubRequest) (pkg.SubscriptionDTO, error) {
	log.Printf("INFO: UpdateSubscription requested for user %s, service %s", userID, serviceName)
	sub, err := s.repo.GetSub(ctx, userID)
	if err != nil {
		log.Printf("ERROR: failed to get subscription for update for user %s, service %s: %v", userID, serviceName, err)
		return pkg.SubscriptionDTO{}, fmt.Errorf("failed to get subscription for update: %w", err)
	}

	if sub.ServiceName != serviceName {
		log.Printf("WARN: Service name mismatch during update for user %s. Expected %s, got %s", userID, serviceName, sub.ServiceName)
		return pkg.SubscriptionDTO{}, en.ErrSubscriptionNotFound
	}

	if req.ServiceName != nil {
		sub.ServiceName = *req.ServiceName
	}
	if req.Price != nil {
		sub.Price = *req.Price
	}
	if req.StartDate != nil {
		sub.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		sub.EndDate = *req.EndDate
	}

	if err := s.repo.UpdateSub(ctx, sub); err != nil {
		log.Printf("ERROR: failed to update subscription for user %s, service %s: %v", userID, serviceName, err)
		return pkg.SubscriptionDTO{}, fmt.Errorf("failed to update subscription: %w", err)
	}

	log.Printf("INFO: Subscription updated successfully for user %s, service %s", userID, serviceName)
	return pkg.SubscriptionDTO{
		UserId:      sub.UserID,
		ServiceName: sub.ServiceName,
		Price:       sub.Price,
		StartDate:   sub.StartDate,
		EndDate:     sub.EndDate,
	}, nil
}

func (s *serviceProvider) DeleteSubscription(ctx context.Context, userID, serviceName string) error {
	log.Printf("INFO: DeleteSubscription requested for user %s, service %s", userID, serviceName)
	sub, err := s.repo.GetSub(ctx, userID)
	if err != nil {
		log.Printf("ERROR: failed to get subscription for delete for user %s, service %s: %v", userID, serviceName, err)
		return fmt.Errorf("failed to get subscription for delete: %w", err)
	}

	if sub.ServiceName != serviceName {
		log.Printf("WARN: Service name mismatch during delete for user %s. Expected %s, got %s", userID, serviceName, sub.ServiceName)
		return en.ErrSubscriptionNotFound
	}

	if err := s.repo.DeleteSub(ctx, userID); err != nil {
		log.Printf("ERROR: failed to delete subscription for user %s, service %s: %v", userID, serviceName, err)
		return fmt.Errorf("failed to delete subscription: %w", err)
	}
	log.Printf("INFO: Subscription deleted successfully for user %s, service %s", userID, serviceName)
	return nil
}

func (s *serviceProvider) GetTotalSubscriptionCost(ctx context.Context, req pkg.GetTotalCostRequest) (pkg.GetTotalCostResponse, error) {
	log.Printf("INFO: GetTotalSubscriptionCost requested for user %s, service %s, from %s to %s", req.UserId, *req.ServiceName, req.StartDate, req.EndDate)

	totalCost, err := s.repo.GetTotalCostByPeriod(ctx, req.UserId, *req.ServiceName, req.StartDate, req.EndDate)
	if err != nil {
		log.Printf("ERROR: failed to get total subscription cost for user %s: %v", req.UserId, err)
		return pkg.GetTotalCostResponse{}, fmt.Errorf("failed to get total subscription cost: %w", err)
	}

	log.Printf("INFO: Total subscription cost for user %s is %d", req.UserId, totalCost)
	return pkg.GetTotalCostResponse{TotalCost: totalCost}, nil
}
