package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
)

// APIRouter defines the usable API routes
func APIRouter(r *mux.Router, store Store) {
	r.Path("/pods").HandlerFunc(driftHandler("/pod", store))
	r.Path("/nodes").HandlerFunc(driftHandler("/node", store))
	r.PathPrefix("/").HandlerFunc(defaultHandler)
}

func driftHandler(keyPrefix string, store Store) func(http.ResponseWriter, *http.Request) {

	fn := func(w http.ResponseWriter, r *http.Request) {

		//vars := mux.Vars(r)

		//fmt.Println("nodePodHandler: %v", vars["event"])
		resp, err := store.GetDriftByKeyPrefix(keyPrefix)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		j, err := json.Marshal(resp)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, err := w.Write([]byte(fmt.Sprintf("JSON Error - %v", err)))
			if err != nil {
				//fmt.Printf("Error cannot write response: %v", err)
			}
		}

		_, err = w.Write(j)
		if err != nil {
			fmt.Printf("Error cannot write response: %v", err)
		}
	}

	return fn
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	msg := fmt.Sprintf("%v - URL: %s", time.Now(), r.URL)
	_, err := w.Write([]byte(msg))
	if err != nil {
		log.Errorf("Error cannot write response: %v", err)
	}
}
