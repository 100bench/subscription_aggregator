package public

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/100bench/subscription_aggregator/internal/entities"
	"github.com/100bench/subscription_aggregator/internal/pkg/validation"
	pkg "github.com/100bench/subscription_aggregator/pkg/dto"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/pkg/errors"
)

type Server struct {
	service PublicService
	router  *chi.Mux
}

func NewServer(service PublicService) (*Server, error) {
	if service == nil {
		return nil, errors.Wrap(entities.ErrNilDependency, "public server service")
	}
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	s := &Server{
		service: service,
		router:  r,
	}
	s.setupRoutes()
	return s, nil
}

func (s *Server) GetRouter() *chi.Mux {
	return s.router
}

func (s *Server) setupRoutes() {
	s.router.Route("/api/v1", func(r chi.Router) {
		r.Post("/subscriptions", s.handleCreateSubscription)
		r.Get("/subscriptions/{id}", s.handleGetSubscription)
		r.Get("/subscriptions", s.handleGetAllSubscriptions)
		r.Put("/subscriptions/{id}", s.handleUpdateSubscription)
		r.Delete("/subscriptions/{id}", s.handleDeleteSubscription)
		r.Get("/subscriptions/calculate", s.handleGetTotalCost)
	})
}

// @Summary Create a new subscription
// @Description Creates a new subscription for a user
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param subscription body pkg.CreateSubRequest true "Subscription"
// @Success 201 {object} pkg.SubscriptionDTO
// @Failure 400 {object} pkg.ErrorResponse
// @Failure 500 {object} pkg.ErrorResponse
// @Router /api/v1/subscriptions [post]
func (s *Server) handleCreateSubscription(w http.ResponseWriter, r *http.Request) {
	var req pkg.CreateSubRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondWithError(w, http.StatusBadRequest, pkg.ErrorResponse{Error: err.Error()})
		return
	}

	// Validate input
	if err := validation.ValidateUUID(req.UserId); err != nil {
		s.respondWithError(w, http.StatusBadRequest, pkg.ErrorResponse{Error: err.Error()})
		return
	}
	if err := validation.ValidateServiceName(req.ServiceName); err != nil {
		s.respondWithError(w, http.StatusBadRequest, pkg.ErrorResponse{Error: err.Error()})
		return
	}
	if err := validation.ValidatePrice(req.Price); err != nil {
		s.respondWithError(w, http.StatusBadRequest, pkg.ErrorResponse{Error: err.Error()})
		return
	}
	if err := validation.ValidateRequiredDate(req.StartDate, "start_date"); err != nil {
		s.respondWithError(w, http.StatusBadRequest, pkg.ErrorResponse{Error: err.Error()})
		return
	}
	if err := validation.ValidateDate(req.EndDate, "end_date"); err != nil {
		s.respondWithError(w, http.StatusBadRequest, pkg.ErrorResponse{Error: err.Error()})
		return
	}
	if err := validation.ValidateDateRange(req.StartDate, req.EndDate); err != nil {
		s.respondWithError(w, http.StatusBadRequest, pkg.ErrorResponse{Error: err.Error()})
		return
	}

	sub := entities.Subscription{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserId,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
	}
	id, err := s.service.CreateSubscription(r.Context(), sub)
	if err != nil {
		s.respondWithError(w, http.StatusInternalServerError, pkg.ErrorResponse{Error: err.Error()})
		return
	}

	// Get the created subscription with ID
	createdSub, err := s.service.GetSubscriptionByID(r.Context(), id)
	if err != nil {
		s.respondWithError(w, http.StatusInternalServerError, pkg.ErrorResponse{Error: err.Error()})
		return
	}

	response := pkg.SubscriptionDTO{
		ID:          createdSub.ID,
		UserId:      createdSub.UserID,
		ServiceName: createdSub.ServiceName,
		Price:       createdSub.Price,
		StartDate:   createdSub.StartDate,
		EndDate:     createdSub.EndDate,
	}
	s.respondWithJSON(w, http.StatusCreated, response)
}

// @Summary Get a subscription
// @Description Get subscription by ID
// @Tags subscriptions
// @Produce json
// @Param id path int true "Subscription ID"
// @Success 200 {object} pkg.SubscriptionDTO
// @Failure 400 {object} pkg.ErrorResponse
// @Failure 404 {object} pkg.ErrorResponse
// @Failure 500 {object} pkg.ErrorResponse
// @Router /api/v1/subscriptions/{id} [get]
func (s *Server) handleGetSubscription(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		s.respondWithError(w, http.StatusBadRequest, pkg.ErrorResponse{Error: "invalid subscription ID"})
		return
	}

	sub, err := s.service.GetSubscriptionByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, entities.ErrSubscriptionNotFound) {
			s.respondWithError(w, http.StatusNotFound, pkg.ErrorResponse{Error: err.Error()})
			return
		}
		s.respondWithError(w, http.StatusInternalServerError, pkg.ErrorResponse{Error: err.Error()})
		return
	}

	resp := pkg.SubscriptionDTO{
		ID:          sub.ID,
		UserId:      sub.UserID,
		ServiceName: sub.ServiceName,
		Price:       sub.Price,
		StartDate:   sub.StartDate,
		EndDate:     sub.EndDate,
	}
	s.respondWithJSON(w, http.StatusOK, resp)
}

// @Summary Get all subscriptions for a user
// @Tags subscriptions
// @Produce json
// @Param user_id query string true "User ID"
// @Param service_name query string false "Service name filter"
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 10, max: 100)"
// @Success 200 {object} pkg.GetSubsResponse
// @Failure 400 {object} pkg.ErrorResponse
// @Failure 500 {object} pkg.ErrorResponse
// @Router /api/v1/subscriptions [get]
func (s *Server) handleGetAllSubscriptions(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if err := validation.ValidateUUID(userID); err != nil {
		s.respondWithError(w, http.StatusBadRequest, pkg.ErrorResponse{Error: err.Error()})
		return
	}

	serviceName := r.URL.Query().Get("service_name")

	// Parse pagination parameters
	page := 1
	pageSize := 10

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	subs, total, err := s.service.GetListSubscriptions(r.Context(), userID, serviceName, page, pageSize)
	if err != nil {
		s.respondWithError(w, http.StatusInternalServerError, pkg.ErrorResponse{Error: err.Error()})
		return
	}

	list := make([]pkg.SubscriptionDTO, 0, len(subs))
	for _, ssub := range subs {
		list = append(list, pkg.SubscriptionDTO{
			ID:          ssub.ID,
			UserId:      ssub.UserID,
			ServiceName: ssub.ServiceName,
			Price:       ssub.Price,
			StartDate:   ssub.StartDate,
			EndDate:     ssub.EndDate,
		})
	}
	s.respondWithJSON(w, http.StatusOK, pkg.GetSubsResponse{
		Subscriptions: list,
		Total:         total,
		Page:          page,
		PageSize:      pageSize,
	})
}

// @Summary Update a subscription
// @Description Update subscription by ID
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path int true "Subscription ID"
// @Param subscription body pkg.UpdateSubRequest true "Update payload"
// @Success 200 {object} pkg.SubscriptionDTO
// @Failure 400 {object} pkg.ErrorResponse
// @Failure 404 {object} pkg.ErrorResponse
// @Failure 500 {object} pkg.ErrorResponse
// @Router /api/v1/subscriptions/{id} [put]
func (s *Server) handleUpdateSubscription(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		s.respondWithError(w, http.StatusBadRequest, pkg.ErrorResponse{Error: "invalid subscription ID"})
		return
	}

	var req pkg.UpdateSubRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondWithError(w, http.StatusBadRequest, pkg.ErrorResponse{Error: err.Error()})
		return
	}

	// Validate input
	if req.Price != nil {
		if err := validation.ValidatePrice(*req.Price); err != nil {
			s.respondWithError(w, http.StatusBadRequest, pkg.ErrorResponse{Error: err.Error()})
			return
		}
	}
	if req.StartDate != nil {
		if err := validation.ValidateDate(*req.StartDate, "start_date"); err != nil {
			s.respondWithError(w, http.StatusBadRequest, pkg.ErrorResponse{Error: err.Error()})
			return
		}
	}
	if req.EndDate != nil {
		if err := validation.ValidateDate(*req.EndDate, "end_date"); err != nil {
			s.respondWithError(w, http.StatusBadRequest, pkg.ErrorResponse{Error: err.Error()})
			return
		}
	}
	// Validate date range if both dates are provided
	if req.StartDate != nil && req.EndDate != nil {
		if err := validation.ValidateDateRange(*req.StartDate, *req.EndDate); err != nil {
			s.respondWithError(w, http.StatusBadRequest, pkg.ErrorResponse{Error: err.Error()})
			return
		}
	}

	if err := s.service.UpdateSubscriptionByID(r.Context(), id, req.Price, req.StartDate, req.EndDate); err != nil {
		if errors.Is(err, entities.ErrSubscriptionNotFound) {
			s.respondWithError(w, http.StatusNotFound, pkg.ErrorResponse{Error: err.Error()})
			return
		}
		s.respondWithError(w, http.StatusInternalServerError, pkg.ErrorResponse{Error: err.Error()})
		return
	}

	// Optionally return the updated subscription state
	sub, err := s.service.GetSubscriptionByID(r.Context(), id)
	if err != nil {
		s.respondWithError(w, http.StatusInternalServerError, pkg.ErrorResponse{Error: err.Error()})
		return
	}
	resp := pkg.SubscriptionDTO{
		ID:          sub.ID,
		UserId:      sub.UserID,
		ServiceName: sub.ServiceName,
		Price:       sub.Price,
		StartDate:   sub.StartDate,
		EndDate:     sub.EndDate,
	}
	s.respondWithJSON(w, http.StatusOK, resp)
}

// @Summary Delete a subscription
// @Description Delete subscription by ID
// @Tags subscriptions
// @Param id path int true "Subscription ID"
// @Success 204
// @Failure 400 {object} pkg.ErrorResponse
// @Failure 404 {object} pkg.ErrorResponse
// @Failure 500 {object} pkg.ErrorResponse
// @Router /api/v1/subscriptions/{id} [delete]
func (s *Server) handleDeleteSubscription(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		s.respondWithError(w, http.StatusBadRequest, pkg.ErrorResponse{Error: "invalid subscription ID"})
		return
	}

	if err := s.service.DeleteSubscriptionByID(r.Context(), id); err != nil {
		if errors.Is(err, entities.ErrSubscriptionNotFound) {
			s.respondWithError(w, http.StatusNotFound, pkg.ErrorResponse{Error: err.Error()})
			return
		}
		s.respondWithError(w, http.StatusInternalServerError, pkg.ErrorResponse{Error: err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Get total cost by period
// @Description Returns total subscription cost for a user in period; service filter optional
// @Tags subscriptions
// @Produce json
// @Param user_id query string true "User ID"
// @Param start_date query string true "Start date MM-YYYY"
// @Param end_date query string true "End date MM-YYYY"
// @Param service_name query string false "Service name"
// @Success 200 {object} pkg.GetTotalCostResponse
// @Failure 400 {object} pkg.ErrorResponse
// @Failure 500 {object} pkg.ErrorResponse
// @Router /api/v1/subscriptions/calculate [get]
func (s *Server) handleGetTotalCost(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	userID := q.Get("user_id")
	startDate := q.Get("start_date")
	endDate := q.Get("end_date")
	serviceName := q.Get("service_name")

	// Validate input
	if err := validation.ValidateUUID(userID); err != nil {
		s.respondWithError(w, http.StatusBadRequest, pkg.ErrorResponse{Error: err.Error()})
		return
	}
	if err := validation.ValidateRequiredDate(startDate, "start_date"); err != nil {
		s.respondWithError(w, http.StatusBadRequest, pkg.ErrorResponse{Error: err.Error()})
		return
	}
	if err := validation.ValidateRequiredDate(endDate, "end_date"); err != nil {
		s.respondWithError(w, http.StatusBadRequest, pkg.ErrorResponse{Error: err.Error()})
		return
	}
	if err := validation.ValidateDateRange(startDate, endDate); err != nil {
		s.respondWithError(w, http.StatusBadRequest, pkg.ErrorResponse{Error: err.Error()})
		return
	}

	total, err := s.service.GetTotalCostByPeriod(r.Context(), userID, serviceName, startDate, endDate)
	if err != nil {
		s.respondWithError(w, http.StatusInternalServerError, pkg.ErrorResponse{Error: err.Error()})
		return
	}

	s.respondWithJSON(w, http.StatusOK, pkg.GetTotalCostResponse{TotalCost: total})
}

func (s *Server) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func (s *Server) respondWithError(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
