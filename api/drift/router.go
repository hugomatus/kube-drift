package drift

import (
	"fmt"
	"github.com/hugomatus/kube-drift/api/store"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
)

// APIRouter defines the usable API routes
func APIRouter(r *mux.Router, s *store.Store) {
	r.Path("/metrics/nodes/{name}").HandlerFunc(cadvisorHandler(s))
	r.Path("/metrics/nodes/{name}/{namespace}").HandlerFunc(cadvisorHandler(s))
	r.Path("/metrics/nodes/{name}/{namespace}/{podname}").HandlerFunc(cadvisorHandler(s))
	r.Path("/metrics/nodes/{name}/{namespace}/{podname}/{metric}").HandlerFunc(cadvisorHandler(s))
	r.Path("/drift/{kind}").HandlerFunc(driftHandler(s))
	r.Path("/drift/{kind}/{namespace}").HandlerFunc(driftHandler(s))
	r.Path("/drift/{kind}/{namespace}/{template-hash}").HandlerFunc(driftHandler(s))
	r.PathPrefix("/").HandlerFunc(defaultHandler)
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	msg := fmt.Sprintf("%v - URL: %s", time.Now(), r.URL)
	_, err := w.Write([]byte(msg))
	if err != nil {
		log.Errorf("Error cannot write response: %v", err)
	}
}
