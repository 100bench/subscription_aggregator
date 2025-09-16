package entities

type Subscription struct {
	ServiceName string
	Price       int
	UserID      string
	StartDate   string
	EndDate     string
}

func NewSubscription(serviceName, userID, startDate, endDate string, price int) (*Subscription, error) {
	return &Subscription{
		ServiceName: serviceName,
		Price:       price,
		UserID:      userID,
		StartDate:   startDate,
		EndDate:     endDate,
	}, nil
}
