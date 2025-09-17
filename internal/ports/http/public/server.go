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
}

func (s *Server) handleCreateSubscription(w http.ResponseWriter, r *http.Request) {
	var req pkg.CreateSubRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondWithError(w, http.StatusBadRequest, pkg.ErrorResponse{Error: err.Error()})
		return
	}

	subDTO, err := s.service.CreateSubscription(r.Context(), req)
	if err != nil {
		s.respondWithError(w, http.StatusInternalServerError, pkg.ErrorResponse{Error: err.Error()})
		return
	}

	s.respondWithJSON(w, http.StatusOK, subDTO)
}

func (s *Server) handleGetSubscription(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	serviceName := chi.URLParam(r, "serviceName")

	subDTO, err := s.service.GetSubscription(r.Context(), userID, serviceName)
	if err != nil {
		if errors.Is(err, entities.ErrSubscriptionNotFound) {
			s.respondWithError(w, http.StatusNotFound, pkg.ErrorResponse{Error: err.Error()})
			return
		}
		s.respondWithError(w, http.StatusInternalServerError, pkg.ErrorResponse{Error: err.Error()})
		return
	}

	s.respondWithJSON(w, http.StatusOK, subDTO)
}

func (s *Server) handleGetAllSubscriptions(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")

	resp, err := s.service.GetAllSubscriptions(r.Context(), userID)
	if err != nil {
		s.respondWithError(w, http.StatusInternalServerError, pkg.ErrorResponse{Error: err.Error()})
		return
	}

	s.respondWithJSON(w, http.StatusOK, resp)
}

func (s *Server) handleUpdateSubscription(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	serviceName := chi.URLParam(r, "serviceName")

	var req pkg.UpdateSubRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondWithError(w, http.StatusBadRequest, pkg.ErrorResponse{Error: err.Error()})
		return
	}

	subDTO, err := s.service.UpdateSubscription(r.Context(), userID, serviceName, req)
	if err != nil {
		if errors.Is(err, entities.ErrSubscriptionNotFound) {
			s.respondWithError(w, http.StatusNotFound, pkg.ErrorResponse{Error: err.Error()})
			return
		}
		s.respondWithError(w, http.StatusInternalServerError, pkg.ErrorResponse{Error: err.Error()})
		return
	}

	s.respondWithJSON(w, http.StatusOK, subDTO)
}

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
