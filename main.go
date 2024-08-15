package main

import (
	"fmt"
	"log"
	"net/http"
)

// struct to hold state of api
type apiConfig struct {
	// fileserverHits holds a int of how many time the server has been accessed
	fileserverHits int
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	// create instance of apiConfig with an intial of 0
	apiCfg := apiConfig{
		fileserverHits: 0,
	}

	// use http multiplexer to route requests to their appropriate handler
	mux := http.NewServeMux()
	// mux.Handle("/app/*", ... ): Registers a handler for paths that start with /app/.
	// The * indicates that anything after /app/ should be handled by this handler.
	// http.StripPrefix("/app", ... ): Strips the /app prefix from the request URL before passing it to the file server.
	// This makes the file server treat /app/somefile as ./somefile.
	// http.FileServer(http.Dir(filepathRoot)): Serves files from the directory specified by filepathRoot. In this case, it's the current directory.
	// apiCfg.middlewareMetricsInc(...): Wraps the file server with middleware that increments the fileserverHits counter every time the file server is accessed.
	mux.Handle("/app/*", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	// checks the servers readiness as in status code
	mux.HandleFunc("/healthz", handlerReadiness)
	// checks the number of server hits
	mux.HandleFunc("/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("/reset", apiCfg.handleReset)

	// creates a new http server with port and mux handler
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}

// func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request): Defines a method on apiConfig that handles requests to the
// /metrics endpoint. It takes two arguments:
//
//	w http.ResponseWriter: Used to write the HTTP response.
//	r *http.Request: Represents the HTTP request.
//
// w.Header().Add("Content-Type", "text/plain; charset=utf-8"): Sets the response content type to plain text with UTF-8 encoding.
// w.WriteHeader(http.StatusOK): Sends a 200 OK status code.
// w.Write([]byte(fmt.Sprintf("Hits: %d", cfg.fileserverHits))): Writes the number of file server hits to the response body.
func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Hits: %d", cfg.fileserverHits)))
}

// func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler: This function is a middleware that wraps another handler.
// It takes a next handler as an argument and returns a new handler that increments the fileserverHits counter.
// return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { ... }): Returns an anonymous function that handles the request.
// cfg.fileserverHits++: Increments the fileserverHits counter every time this middleware is invoked.
// next.ServeHTTP(w, r): Passes the request to the next handler in the chain.
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits++
		next.ServeHTTP(w, r)
	})
}

// writes a status code to check on server health
func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (cfg *apiConfig) handleReset(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	cfg.fileserverHits = 0
	w.Write([]byte("Hits reset to 0"))
}
