package main

import (
	"encoding/json"
	"fmt"
	"housing-anywhere/models"
	"housing-anywhere/services"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/gorilla/mux"
)

func navigation(w http.ResponseWriter, r *http.Request) {
	log.Println("jhgtck")
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	var location models.Location
	err = json.Unmarshal(body, &location)
	if err != nil {
		http.Error(w, "Error parsing body content", http.StatusBadRequest)
		return
	}

	loc, err := services.Calculate(location)
	if err != nil {
		http.Error(w, "Error calculating location", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	io.WriteString(w, `{"loc": `+fmt.Sprintf("%.2f", loc)+`}`)
}

func healthcheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	io.WriteString(w, `{"alive": true}`)
}
func main() {
	r := mux.NewRouter()
	lmt := tollbooth.NewLimiter(3, &limiter.ExpirableOptions{DefaultExpirationTTL: time.Hour})

	var md mux.MiddlewareFunc
	md = func(r http.Handler) http.Handler {
		lmtmw := tollbooth.LimitHandler(lmt, r)
		return lmtmw
	}
	r.Use(md)
	api := r.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/dns", navigation).Methods(http.MethodPost)
	api.HandleFunc("/health", healthcheck).Methods(http.MethodGet)

	log.Print("Listening on localhost:5000")
	log.Fatal(http.ListenAndServe(":8080", r))
}
