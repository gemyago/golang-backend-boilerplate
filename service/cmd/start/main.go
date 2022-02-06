package main

import (
	"evgeny-myasishchev/golang-boilerplate/service/pkg/api/healthcheckv1"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	// HC routes
	r.Mount("/v1/hc/", healthcheckv1.NewRoutes())

	log.Print("Starting server on port 3000")
	srv := newHttpServer("localhost:3000", r.ServeHTTP)
	srv.ListenAndServe()
}

func newHttpServer(addr string, handler http.HandlerFunc) *http.Server {
	return &http.Server{
		Addr:         addr,
		IdleTimeout:  7 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 5 * time.Second,
		Handler:      handler,
	}
}
