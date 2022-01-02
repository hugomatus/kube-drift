package metrics

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hugomatus/kube-drift/api/store"
	appLog "k8s.io/klog/v2"
	"net/http"
	"time"
)

// APIRouter defines the usable API routes
func APIRouter(r *mux.Router, s *store.Store) {
	r.Path("/{name}").HandlerFunc(metricsHandler(s))
	r.Path("/{name}/{namespace}").HandlerFunc(metricsHandler(s))
	r.Path("/{name}/{namespace}/{podname}").HandlerFunc(metricsHandler(s))
	r.Path("/{name}/{namespace}/{podname}/{container}").HandlerFunc(metricsHandler(s))
	r.Path("/{name}/{namespace}/{podname}/{container}/{metric}").HandlerFunc(metricsHandler(s))
	r.PathPrefix("/").HandlerFunc(defaultHandler)
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	msg := fmt.Sprintf("%v - URL: %s", time.Now(), r.URL)
	_, err := w.Write([]byte(msg))
	if err != nil {
		appLog.Errorf("Error cannot write response: %v", err)
	}
}
