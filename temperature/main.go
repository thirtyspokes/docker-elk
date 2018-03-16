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
)

var log *logrus.Logger

type contextKey string

// TemperatureResponse represents the response body of a hypothetical
// temperature service that accepts a latitude and longitude and returns
// temperature data.
type TemperatureResponse struct {
	Latitude              string  `json:"latitude"`
	Longitude             string  `json:"longitude"`
	TemperatureCelsius    float64 `json:"temp_celsius"`
	TemperatureFahrenheit float64 `json:"temp_fahrenheit"`
}

func (c contextKey) String() string {
	return "temperature context key " + string(c)
}

func main() {
	var conn net.Conn
	var err error

	// We don't want to start up until logstash is actually running, so
	// we will retry connections until success before moving on.  You definitely
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
	hook, err := logrustash.NewHookWithConn(conn, "temperature")
	if err != nil {
		log.Fatal(err)
	}
	log.Hooks.Add(hook)

	http.Handle("/temperature", middleware(http.HandlerFunc(handler)))
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	log := r.Context().Value(contextKey("logger")).(*logrus.Entry)

	latitude := r.URL.Query().Get("latitude")
	if latitude == "" {
		http.Error(w, "a non-empty latitude must be supplied", http.StatusBadRequest)
		return
	}

	longitude := r.URL.Query().Get("longitude")
	if longitude == "" {
		http.Error(w, "a non-empty latitude must be supplied", http.StatusBadRequest)
		return
	}

	log.Infof("Query received: %s, %s", latitude, longitude)

	response := &TemperatureResponse{
		Latitude:              latitude,
		Longitude:             longitude,
		TemperatureCelsius:    21.111,
		TemperatureFahrenheit: 70.0,
	}

	log.Infof("Temperature captured for location: %s, %s: %f degrees fahrenheit", latitude, longitude, response.TemperatureFahrenheit)

	json, err := json.Marshal(response)
	if err != nil {
		log.Errorf("Failed to serialize temperature response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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
