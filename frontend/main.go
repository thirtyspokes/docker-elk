package main

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"time"

	"github.com/bshuster-repo/logrus-logstash-hook"
	"github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"github.com/thirtyspokes/docker-elk/frontend/requests"
)

var log *logrus.Logger

type contextKey string

func (c contextKey) String() string {
	return "frontend context key " + string(c)
}

func main() {
	var conn net.Conn
	var err error

	// We don't want to start up until logstash is actually running, so
	// we will retry connections until success before moving on.  You probably
	// shouldn't do this in production.
	for {
		conn, err = net.Dial("tcp", "logstash:5000")
		if err != nil {
			time.Sleep(5000 * time.Millisecond)
		} else {
			break
		}
	}

	log = logrus.New()
	hook, err := logrustash.NewHookWithConn(conn, "frontend")
	if err != nil {
		log.Fatal(err)
	}
	log.Hooks.Add(hook)

	http.Handle("/query", middleware(http.HandlerFunc(handler)))
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	log := r.Context().Value(contextKey("logger")).(*logrus.Entry)
	requestID := r.Context().Value(contextKey("request-id")).(string)

	// Grab the requested street address from the client
	location := r.URL.Query().Get("address")
	if location == "" {
		http.Error(w, "a non-empty address must be supplied", http.StatusBadRequest)
		return
	}

	// Send the address to the geolocation service.
	log.Infof("Sending address for geolocation: %s", location)
	locationResponse, err := requests.GeolocateAddress(requestID, location)
	if err != nil {
		log.Errorf("Failed to retrieve a valid address from location service: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Use the geolocation service's response to query for the temperature.
	log.Infof("Querying for temperature at %s, %s", locationResponse.Latitude, locationResponse.Longitude)
	tempResponse, err := requests.GetTemperatureByLocation(requestID, locationResponse.Latitude, locationResponse.Longitude)
	if err != nil {
		log.Errorf("Failed to retrieve temperature data from temperature service: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// If all is well, return the request information.
	response := &queryResponse{
		Address:        location,
		Latitude:       locationResponse.Latitude,
		Longitude:      locationResponse.Longitude,
		TempCelcius:    tempResponse.TemperatureCelsius,
		TempFahrenheit: tempResponse.TemperatureFahrenheit,
	}

	json, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Write(json)
	return
}

func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		// Set the request ID, or generate one if needed
		id := req.Header.Get("X-Request-Id")
		if id == "" {
			id = uuid.NewV4().String()
		}

		// Make a new log entry with the request-id and path set
		entry := log.WithFields(logrus.Fields{
			"request-id": id,
			"endpoint":   req.URL.Path,
			"method":     req.Method,
		})

		// Add the ID and logger to the context
		ctx := context.WithValue(req.Context(), contextKey("request-id"), id)
		ctx = context.WithValue(ctx, contextKey("logger"), entry)

		// Set it in the response headers, so we don't have to remember to do that
		rw.Header().Set("X-Request-Id", id)

		// Set our content-type as well
		rw.Header().Set("Content-Type", "application/json")

		next.ServeHTTP(rw, req.WithContext(ctx))

	})
}

type queryResponse struct {
	Address        string  `json:"address"`
	Latitude       string  `json:"latitude"`
	Longitude      string  `json:"longitude"`
	TempCelcius    float64 `json:"temp_celsius"`
	TempFahrenheit float64 `json:"temp_fahrenheit"`
}
