package provider

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
)

// DashboardRouter defines the usable API routes
func DashboardRouter(r *mux.Router, db *sql.DB) {
	r.Path("/node").HandlerFunc(nodeHandler(db))
	r.Path("/namespaces/{Namespace}}").HandlerFunc(podHandler(db))
	r.PathPrefix("/").HandlerFunc(defaultHandler)
}

func podHandler(db *sql.DB) func(http.ResponseWriter, *http.Request) {

	return nil
}

func nodeHandler(db *sql.DB) func(http.ResponseWriter, *http.Request) {

	return nil
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	msg := fmt.Sprintf("%v - URL: %s", time.Now(), r.URL)
	_, err := w.Write([]byte(msg))
	if err != nil {
		log.Errorf("Error cannot write response: %v", err)
	}
}
