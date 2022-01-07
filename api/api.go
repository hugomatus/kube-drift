package api

import (
	"fmt"
	"github.com/hugomatus/kube-drift/api/drift"
	"github.com/hugomatus/kube-drift/api/metrics"
	"github.com/hugomatus/kube-drift/api/store"
	appLog "k8s.io/klog/v2"
	"net/http"

	"github.com/gorilla/mux"
)

// Manager provides a handler for all api calls
func Manager(r *mux.Router, s *store.Store) {
	routerDrift := r.PathPrefix("/api/v1/drift").Subrouter()
	drift.APIRouter(routerDrift, s)
	routerMetrics := r.PathPrefix("/api/v1/metrics").Subrouter()
	metrics.APIRouter(routerMetrics, s)
	r.PathPrefix("/").HandlerFunc(DefaultHandler)
}

// DefaultHandler provides a handler for all http calls
func DefaultHandler(w http.ResponseWriter, r *http.Request) {
	msg := fmt.Sprintf("URL: %s", r.URL)
	appLog.Infof(msg)
	_, err := w.Write([]byte(msg))
	if err != nil {
		appLog.Errorf("Error cannot write response: %v", err)
	}
}
