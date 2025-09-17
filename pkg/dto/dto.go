package pkg

type SubscriptionDTO struct {
	UserId      string `json:"user_id" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	ServiceName string `json:"service_name" example:"Yandex Plus"`
	Price       int    `json:"price" example:"400"`
	StartDate   string `json:"start_date" example:"07-2025"`
	EndDate     string `json:"end_date,omitempty" example:"07-2026"`
}

type GetSubsResponse struct {
	Subscriptions []SubscriptionDTO `json:"subscriptions"`
}

type UpdateSubRequest struct {
	ServiceName *string `json:"service_name,omitempty" example:"Netflix"`
	Price       *int    `json:"price,omitempty" example:"500"`
	StartDate   *string `json:"start_date,omitempty" example:"08-2025"`
	EndDate     *string `json:"end_date,omitempty" example:"08-2026"`
}

type CreateSubRequest struct {
	UserId      string `json:"user_id" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	ServiceName string `json:"service_name" example:"Yandex Plus"`
	Price       int    `json:"price" example:"400"`
	StartDate   string `json:"start_date" example:"07-2025"`
	EndDate     string `json:"end_date,omitempty" example:"07-2026"`
}

type DeleteSubRequest struct {
	UserId      string `json:"user_id" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	ServiceName string `json:"service_name" example:"Yandex Plus"`
}

type GetSubsRequest struct {
	UserId string `json:"user_id" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"Subscription not found"`
}

type GetTotalCostRequest struct {
	UserId      string  `json:"user_id" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	ServiceName *string `json:"service_name,omitempty" example:"Netflix"`
	StartDate   string  `json:"start_date" example:"01-2025"`
	EndDate     string  `json:"end_date" example:"12-2025"`
}

type GetTotalCostResponse struct {
	TotalCost int `json:"total_cost" example:"1200"`
}
