# agregator

## entities
type subscription struct{
    ServiceName string
    Price int
    UserID string
    StartDate string
    EndDate string
}

type errors struct{
    ErrNilDependency
}

## cases

## interfaces
### DB
type SubscriptionRepository interface {
    CreateSub(subscription Subscription) error
    GetSub(userID string) (Subscription, error)
    UpdateSub(subscription Subscription) error
    DeleteSub(userID string) error
    GetListSubs(userId string) ([]Subscription, error)
}
### provider
type SubProviderClient interface {
    FetchSub(userID string) (Subscription, error)
}
### consumer
type SubConsumerClient interface {
    SendSub(subscription Subscription) error
}
## adapters
### provider
отдает данные о подписках в
### consumer

## ports
### Chi.Mux handlers