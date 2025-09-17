package public

import (
	"encoding/json"
	"net/http"

	"github.com/100bench/subscription_aggregator/internal/entities"
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
	s.router.Post("/subscriptions", s.handleCreateSubscription)
	s.router.Get("/subscriptions/{userID}/{serviceName}", s.handleGetSubscription)
	s.router.Get("/subscriptions/{userID}", s.handleGetAllSubscriptions)
	s.router.Put("/subscriptions/{userID}/{serviceName}", s.handleUpdateSubscription)
	s.router.Delete("/subscriptions/{userID}/{serviceName}", s.handleDeleteSubscription)
	s.router.Get("/subscriptions/total-cost", s.handleGetTotalCost)
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
// @Router /subscriptions [post]
func (s *Server) handleCreateSubscription(w http.ResponseWriter, r *http.Request) {
	var req pkg.CreateSubRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
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
	if err := s.service.CreateSubscription(r.Context(), sub); err != nil {
		s.respondWithError(w, http.StatusInternalServerError, pkg.ErrorResponse{Error: err.Error()})
		return
	}

	created := pkg.SubscriptionDTO(req)
	s.respondWithJSON(w, http.StatusCreated, created)
}

// @Summary Get a subscription
// @Description Get subscription by user ID and service name
// @Tags subscriptions
// @Produce json
// @Param userID path string true "User ID"
// @Param serviceName path string true "Service Name"
// @Success 200 {object} pkg.SubscriptionDTO
// @Failure 404 {object} pkg.ErrorResponse
// @Failure 500 {object} pkg.ErrorResponse
// @Router /subscriptions/{userID}/{serviceName} [get]
func (s *Server) handleGetSubscription(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	serviceName := chi.URLParam(r, "serviceName")

	sub, err := s.service.GetSubscription(r.Context(), userID, serviceName)
	if err != nil {
		if errors.Is(err, entities.ErrSubscriptionNotFound) {
			s.respondWithError(w, http.StatusNotFound, pkg.ErrorResponse{Error: err.Error()})
			return
		}
		s.respondWithError(w, http.StatusInternalServerError, pkg.ErrorResponse{Error: err.Error()})
		return
	}

	resp := pkg.SubscriptionDTO{
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
// @Param userID path string true "User ID"
// @Success 200 {object} pkg.GetSubsResponse
// @Failure 500 {object} pkg.ErrorResponse
// @Router /subscriptions/{userID} [get]
func (s *Server) handleGetAllSubscriptions(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")

	subs, err := s.service.GetListSubscriptions(r.Context(), userID)
	if err != nil {
		s.respondWithError(w, http.StatusInternalServerError, pkg.ErrorResponse{Error: err.Error()})
		return
	}

	list := make([]pkg.SubscriptionDTO, 0, len(subs))
	for _, ssub := range subs {
		list = append(list, pkg.SubscriptionDTO{
			UserId:      ssub.UserID,
			ServiceName: ssub.ServiceName,
			Price:       ssub.Price,
			StartDate:   ssub.StartDate,
			EndDate:     ssub.EndDate,
		})
	}
	s.respondWithJSON(w, http.StatusOK, pkg.GetSubsResponse{Subscriptions: list})
}

// @Summary Update a subscription
// @Description Update by user ID and service name
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param userID path string true "User ID"
// @Param serviceName path string true "Service Name"
// @Param subscription body pkg.UpdateSubRequest true "Update payload"
// @Success 200 {object} pkg.SubscriptionDTO
// @Failure 400 {object} pkg.ErrorResponse
// @Failure 404 {object} pkg.ErrorResponse
// @Failure 500 {object} pkg.ErrorResponse
// @Router /subscriptions/{userID}/{serviceName} [put]
func (s *Server) handleUpdateSubscription(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	serviceName := chi.URLParam(r, "serviceName")

	var req pkg.UpdateSubRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondWithError(w, http.StatusBadRequest, pkg.ErrorResponse{Error: err.Error()})
		return
	}

	if err := s.service.UpdateSubscription(r.Context(), userID, serviceName, req.Price, req.StartDate, req.EndDate); err != nil {
		if errors.Is(err, entities.ErrSubscriptionNotFound) {
			s.respondWithError(w, http.StatusNotFound, pkg.ErrorResponse{Error: err.Error()})
			return
		}
		s.respondWithError(w, http.StatusInternalServerError, pkg.ErrorResponse{Error: err.Error()})
		return
	}

	// Optionally return the updated subscription state
	sub, err := s.service.GetSubscription(r.Context(), userID, serviceName)
	if err != nil {
		s.respondWithError(w, http.StatusInternalServerError, pkg.ErrorResponse{Error: err.Error()})
		return
	}
	resp := pkg.SubscriptionDTO{
		UserId:      sub.UserID,
		ServiceName: sub.ServiceName,
		Price:       sub.Price,
		StartDate:   sub.StartDate,
		EndDate:     sub.EndDate,
	}
	s.respondWithJSON(w, http.StatusOK, resp)
}

// @Summary Delete a subscription
// @Tags subscriptions
// @Param userID path string true "User ID"
// @Param serviceName path string true "Service Name"
// @Success 204
// @Failure 404 {object} pkg.ErrorResponse
// @Failure 500 {object} pkg.ErrorResponse
// @Router /subscriptions/{userID}/{serviceName} [delete]
func (s *Server) handleDeleteSubscription(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	serviceName := chi.URLParam(r, "serviceName")

	if err := s.service.DeleteSubscription(r.Context(), userID, serviceName); err != nil {
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
// @Router /subscriptions/total-cost [get]
func (s *Server) handleGetTotalCost(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	userID := q.Get("user_id")
	startDate := q.Get("start_date")
	endDate := q.Get("end_date")
	serviceName := q.Get("service_name")

	if userID == "" || startDate == "" || endDate == "" {
		s.respondWithError(w, http.StatusBadRequest, pkg.ErrorResponse{Error: "missing required query params: user_id, start_date, end_date"})
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
