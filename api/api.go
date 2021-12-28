package api

import (
	"fmt"
	"github.com/hugomatus/kube-drift/api/drift"
	"github.com/hugomatus/kube-drift/api/store"
	appLog "k8s.io/klog/v2"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

// Manager provides a handler for all api calls
func Manager(r *mux.Router, store *store.Store) {
	router := r.PathPrefix("/api/v1").Subrouter()
	drift.APIRouter(router, store)
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
