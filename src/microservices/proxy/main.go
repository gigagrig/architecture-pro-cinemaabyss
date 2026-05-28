package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type config struct {
	port                   string
	monolithURL            *url.URL
	moviesServiceURL       *url.URL
	eventsServiceURL       *url.URL
	gradualMigration       bool
	moviesMigrationPercent int
}

func main() {
	rand.Seed(time.Now().UnixNano())

	cfg := loadConfig()
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		target := pickTarget(cfg, r.URL.Path)
		log.Printf("proxying %s %s to %s", r.Method, r.URL.RequestURI(), target.String())
		newReverseProxy(target).ServeHTTP(w, r)
	}))

	log.Printf("starting proxy service on port %s", cfg.port)
	log.Fatal(http.ListenAndServe(":"+cfg.port, mux))
}

func loadConfig() config {
	port := envOrDefault("PORT", "8000")

	return config{
		port:                   port,
		monolithURL:            mustParseURL(envOrDefault("MONOLITH_URL", "http://localhost:8080")),
		moviesServiceURL:       mustParseURL(envOrDefault("MOVIES_SERVICE_URL", "http://localhost:8081")),
		eventsServiceURL:       mustParseURL(envOrDefault("EVENTS_SERVICE_URL", "http://localhost:8082")),
		gradualMigration:       strings.EqualFold(envOrDefault("GRADUAL_MIGRATION", "false"), "true"),
		moviesMigrationPercent: clampPercent(envOrDefault("MOVIES_MIGRATION_PERCENT", "0")),
	}
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Strangler Fig Proxy is healthy"))
}

func pickTarget(cfg config, path string) *url.URL {
	switch {
	case strings.HasPrefix(path, "/api/events"):
		return cfg.eventsServiceURL
	case strings.HasPrefix(path, "/api/movies"):
		if !cfg.gradualMigration {
			return cfg.moviesServiceURL
		}
		if rand.Intn(100) < cfg.moviesMigrationPercent {
			return cfg.moviesServiceURL
		}
		return cfg.monolithURL
	default:
		return cfg.monolithURL
	}
}

func newReverseProxy(target *url.URL) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(target)
	originalDirector := proxy.Director
	proxy.Director = func(r *http.Request) {
		originalDirector(r)
		r.Host = target.Host
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, _ *http.Request, err error) {
		log.Printf("proxy error for %s: %v", target.String(), err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "upstream service is unavailable",
		})
	}
	return proxy
}

func mustParseURL(raw string) *url.URL {
	parsed, err := url.Parse(raw)
	if err != nil {
		log.Fatalf("invalid url %q: %v", raw, err)
	}
	return parsed
}

func clampPercent(raw string) int {
	percent, err := strconv.Atoi(raw)
	if err != nil {
		log.Printf("invalid MOVIES_MIGRATION_PERCENT=%q, using 0", raw)
		return 0
	}
	if percent < 0 {
		return 0
	}
	if percent > 100 {
		return 100
	}
	return percent
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
