package provider

import (
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
)

// APIRouter defines the usable API routes
func APIRouter(r *mux.Router, store *Store) {
	r.Path("/{kind}").HandlerFunc(driftHandler(store))
	r.Path("/{kind}/{namespace}").HandlerFunc(driftHandler(store))
	r.Path("/{kind}/{namespace}/{template-hash}").HandlerFunc(driftHandler(store))
	r.Path("/nodes/metrics/stats/summary").HandlerFunc(statsSummaryHandler(store))
	r.Path("/nodes/{Name}/metrics/stats/summary").HandlerFunc(statsSummaryHandler(store))
	r.PathPrefix("/").HandlerFunc(defaultHandler)
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	msg := fmt.Sprintf("%v - URL: %s", time.Now(), r.URL)
	_, err := w.Write([]byte(msg))
	if err != nil {
		log.Errorf("Error cannot write response: %v", err)
	}
}
