package nuke

import (
	"context"
	"errors"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

type CORSOptions struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

func DefaultCORSOptions() *CORSOptions {
	return &CORSOptions{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		ExposedHeaders:   []string{},
		AllowCredentials: false,
		MaxAge:           86400,
	}
}

func CORS(options *CORSOptions) func(http.Handler) http.Handler {
	if options == nil {
		options = DefaultCORSOptions()
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if len(options.AllowedOrigins) == 1 && options.AllowedOrigins[0] == "*" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if slices.Contains(options.AllowedOrigins, origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			if options.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			if r.Method == "OPTIONS" {
				if len(options.AllowedMethods) > 0 {
					w.Header().Set("Access-Control-Allow-Methods", strings.Join(options.AllowedMethods, ", "))
				}
				if len(options.AllowedHeaders) > 0 {
					w.Header().Set("Access-Control-Allow-Headers", strings.Join(options.AllowedHeaders, ", "))
				}
				if len(options.ExposedHeaders) > 0 {
					w.Header().Set("Access-Control-Expose-Headers", strings.Join(options.ExposedHeaders, ", "))
				}
				if options.MaxAge > 0 {
					w.Header().Set("Access-Control-Max-Age", string(rune(options.MaxAge)))
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func Heartbeat(endpoint string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if (r.Method == "GET" || r.Method == "HEAD") && strings.EqualFold(r.URL.Path, endpoint) {
				w.Header().Set(contentType, applicationJSON)
				w.Header().Set(contentLength, strconv.Itoa(len(healthResponse)))
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(healthResponse))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func Timeout(duration time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), duration)
			defer cancel()
			r = r.WithContext(ctx)
			done := make(chan struct{})
			var once sync.Once
			go func() {
				next.ServeHTTP(w, r)
				once.Do(func() { close(done) })
			}()
			select {
			case <-ctx.Done():
				if errors.Is(ctx.Err(), context.DeadlineExceeded) {
					once.Do(func() {
						w.Header().Set(contentType, applicationJSON)
						w.Header().Set(contentLength, strconv.Itoa(len(timeoutResponse)))
						w.WriteHeader(http.StatusGatewayTimeout)
						w.Write([]byte(timeoutResponse))
					})
				}
			case <-done:
			}
		})
	}
}

func Recover() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("Recovered panic from request handler : %v", err)
					w.WriteHeader(http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
