package public

import (
	"encoding/json"
	"net/http"

	"github.com/100bench/subscription_aggregator/internal/cases"
	"github.com/100bench/subscription_aggregator/internal/entities"
	pkg "github.com/100bench/subscription_aggregator/pkg/dto"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

type HttpServer struct {
	subscriptionService cases.SubscriptionService
	router              *mux.Router
}

func NewHttpServer(subscriptionService cases.SubscriptionService) *HttpServer {
	r := mux.NewRouter()
	s := &HttpServer{
		subscriptionService: subscriptionService,
		router:              r,
	}
	s.setupRoutes()
	return s
}

func (s *HttpServer) GetRouter() *mux.Router {
	return s.router
}

func (s *HttpServer) setupRoutes() {
	s.router.HandleFunc("/subscriptions", s.handleCreateSubscription()).Methods("POST")
	s.router.HandleFunc("/subscriptions/{userID}/{serviceName}", s.handleGetSubscription()).Methods("GET")
	s.router.HandleFunc("/subscriptions/{userID}", s.handleGetAllSubscriptions()).Methods("GET")
	s.router.HandleFunc("/subscriptions/{userID}/{serviceName}", s.handleUpdateSubscription()).Methods("PUT")
	s.router.HandleFunc("/subscriptions/{userID}/{serviceName}", s.handleDeleteSubscription()).Methods("DELETE")
}

func (s *HttpServer) handleCreateSubscription() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req pkg.CreateSubRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		subDTO, err := s.subscriptionService.CreateSubscription(r.Context(), req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(subDTO)
	}
}

func (s *HttpServer) handleGetSubscription() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		userID := vars["userID"]
		serviceName := vars["serviceName"]

		subDTO, err := s.subscriptionService.GetSubscription(r.Context(), userID, serviceName)
		if err != nil {
			if errors.Is(err, entities.ErrSubscriptionNotFound) {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(subDTO)
	}
}

func (s *HttpServer) handleGetAllSubscriptions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		userID := vars["userID"]

		resp, err := s.subscriptionService.GetAllSubscriptions(r.Context(), userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func (s *HttpServer) handleUpdateSubscription() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		userID := vars["userID"]
		serviceName := vars["serviceName"]

		var req pkg.UpdateSubRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		subDTO, err := s.subscriptionService.UpdateSubscription(r.Context(), userID, serviceName, req)
		if err != nil {
			if errors.Is(err, entities.ErrSubscriptionNotFound) {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(subDTO)
	}
}

func (s *HttpServer) handleDeleteSubscription() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		userID := vars["userID"]
		serviceName := vars["serviceName"]

		if err := s.subscriptionService.DeleteSubscription(r.Context(), userID, serviceName); err != nil {
			if errors.Is(err, entities.ErrSubscriptionNotFound) {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
