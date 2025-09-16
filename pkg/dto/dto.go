package pkg

type SubscriptionDTO struct {
	UserId      string `json:"user_id"`
	ServiceName string `json:"service_name"`
	Price       int    `json:"price"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date,omitempty"`
}

type GetSubsResponse struct {
	Subscriptions []SubscriptionDTO `json:"subscriptions"`
}

type UpdateSubRequest struct {
	ServiceName *string `json:"service_name,omitempty"`
	Price       *int    `json:"price,omitempty"`
	StartDate   *string `json:"start_date,omitempty"`
	EndDate     *string `json:"end_date,omitempty"`
}

type CreateSubRequest struct {
	UserId      string `json:"user_id"`
	ServiceName string `json:"service_name"`
	Price       int    `json:"price"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date,omitempty"`
}

type DeleteSubRequest struct {
	UserId      string `json:"user_id"`
	ServiceName string `json:"service_name"`
}

type GetSubsRequest struct {
	UserId string `json:"user_id"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type GetTotalCostRequest struct {
	UserId      string  `json:"user_id"`
	ServiceName *string `json:"service_name,omitempty"` // Optional filter
	StartDate   string  `json:"start_date"`             // Format: YYYY-MM
	EndDate     string  `json:"end_date"`               // Format: YYYY-MM
}

type GetTotalCostResponse struct {
	TotalCost int `json:"total_cost"`
}
