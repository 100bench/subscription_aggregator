package public

import (
	"encoding/json"
	"log"
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
	s.router.HandleFunc("/subscriptions/{userID}/total_cost", s.handleGetTotalSubscriptionCost()).Methods("GET")
}

func (s *HttpServer) handleCreateSubscription() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("INFO: [HTTP] Received CreateSubscription request from %s", r.RemoteAddr)
		var req pkg.CreateSubRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("WARN: [HTTP] Bad request for CreateSubscription from %s: %v", r.RemoteAddr, err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		subDTO, err := s.subscriptionService.CreateSubscription(r.Context(), req)
		if err != nil {
			log.Printf("ERROR: [HTTP] Failed to create subscription for user %s, service %s: %v", req.UserId, req.ServiceName, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(subDTO); err != nil {
			log.Printf("ERROR: [HTTP] Failed to write response for CreateSubscription: %v", err)
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
		log.Printf("INFO: [HTTP] Created subscription for user %s, service %s. Response sent.", subDTO.UserId, subDTO.ServiceName)
	}
}

func (s *HttpServer) handleGetSubscription() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		userID := vars["userID"]
		serviceName := vars["serviceName"]
		log.Printf("INFO: [HTTP] Received GetSubscription request for user %s, service %s from %s", userID, serviceName, r.RemoteAddr)

		subDTO, err := s.subscriptionService.GetSubscription(r.Context(), userID, serviceName)
		if err != nil {
			if errors.Is(err, entities.ErrSubscriptionNotFound) {
				log.Printf("WARN: [HTTP] Subscription not found for user %s, service %s: %v", userID, serviceName, err)
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			log.Printf("ERROR: [HTTP] Failed to get subscription for user %s, service %s: %v", userID, serviceName, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(subDTO); err != nil {
			log.Printf("ERROR: [HTTP] Failed to write response for GetSubscription: %v", err)
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
		log.Printf("INFO: [HTTP] Retrieved subscription for user %s, service %s. Response sent.", userID, serviceName)
	}
}

func (s *HttpServer) handleGetAllSubscriptions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		userID := vars["userID"]
		log.Printf("INFO: [HTTP] Received GetAllSubscriptions request for user %s from %s", userID, r.RemoteAddr)

		resp, err := s.subscriptionService.GetAllSubscriptions(r.Context(), userID)
		if err != nil {
			log.Printf("ERROR: [HTTP] Failed to get all subscriptions for user %s: %v", userID, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("ERROR: [HTTP] Failed to write response for GetAllSubscriptions: %v", err)
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
		log.Printf("INFO: [HTTP] Retrieved all %d subscriptions for user %s. Response sent.", len(resp.Subscriptions), userID)
	}
}

func (s *HttpServer) handleUpdateSubscription() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		userID := vars["userID"]
		serviceName := vars["serviceName"]
		log.Printf("INFO: [HTTP] Received UpdateSubscription request for user %s, service %s from %s", userID, serviceName, r.RemoteAddr)

		var req pkg.UpdateSubRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("WARN: [HTTP] Bad request for UpdateSubscription from %s: %v", r.RemoteAddr, err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		subDTO, err := s.subscriptionService.UpdateSubscription(r.Context(), userID, serviceName, req)
		if err != nil {
			if errors.Is(err, entities.ErrSubscriptionNotFound) {
				log.Printf("WARN: [HTTP] Subscription not found for update for user %s, service %s: %v", userID, serviceName, err)
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			log.Printf("ERROR: [HTTP] Failed to update subscription for user %s, service %s: %v", userID, serviceName, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(subDTO); err != nil {
			log.Printf("ERROR: [HTTP] Failed to write response for UpdateSubscription: %v", err)
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
		log.Printf("INFO: [HTTP] Updated subscription for user %s, service %s. Response sent.", userID, serviceName)
	}
}

func (s *HttpServer) handleDeleteSubscription() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		userID := vars["userID"]
		serviceName := vars["serviceName"]
		log.Printf("INFO: [HTTP] Received DeleteSubscription request for user %s, service %s from %s", userID, serviceName, r.RemoteAddr)

		if err := s.subscriptionService.DeleteSubscription(r.Context(), userID, serviceName); err != nil {
			if errors.Is(err, entities.ErrSubscriptionNotFound) {
				log.Printf("WARN: [HTTP] Subscription not found for delete for user %s, service %s: %v", userID, serviceName, err)
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			log.Printf("ERROR: [HTTP] Failed to delete subscription for user %s, service %s: %v", userID, serviceName, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
		log.Printf("INFO: [HTTP] Deleted subscription for user %s, service %s. No content response sent.", userID, serviceName)
	}
}

func (s *HttpServer) handleGetTotalSubscriptionCost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		userID := vars["userID"]
		startDate := r.URL.Query().Get("start_date")
		endDate := r.URL.Query().Get("end_date")
		serviceName := r.URL.Query().Get("service_name") // Optional

		if userID == "" || startDate == "" || endDate == "" {
			log.Printf("WARN: [HTTP] Bad request for GetTotalSubscriptionCost from %s: missing userID, start_date or end_date", r.RemoteAddr)
			http.Error(w, "missing userID, start_date or end_date", http.StatusBadRequest)
			return
		}

		req := pkg.GetTotalCostRequest{
			UserId:    userID,
			StartDate: startDate,
			EndDate:   endDate,
		}

		if serviceName != "" {
			req.ServiceName = &serviceName
		}

		log.Printf("INFO: [HTTP] Received GetTotalSubscriptionCost request for user %s, service %s, from %s to %s from %s", userID, serviceName, startDate, endDate, r.RemoteAddr)

		resp, err := s.subscriptionService.GetTotalSubscriptionCost(r.Context(), req)
		if err != nil {
			log.Printf("ERROR: [HTTP] Failed to get total subscription cost for user %s: %v", userID, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("ERROR: [HTTP] Failed to write response for GetTotalSubscriptionCost: %v", err)
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
		log.Printf("INFO: [HTTP] Total subscription cost for user %s is %d. Response sent.", userID, resp.TotalCost)
	}
}
