package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/bshuster-repo/logrus-logstash-hook"
	"github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
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

	http.Handle("/temperature", middleware(http.HandlerFunc(handler)))
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	log := r.Context().Value(contextKey("logger")).(*logrus.Entry)

	location := r.URL.Query().Get("address")
	if location == "" {
		http.Error(w, "a non-empty address must be supplied", http.StatusBadRequest)
		return
	}

	log.Info(fmt.Sprintf("Query received: %s", location))
	fmt.Fprintf(w, "hello, world")
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

		next.ServeHTTP(rw, req.WithContext(ctx))

	})
}
