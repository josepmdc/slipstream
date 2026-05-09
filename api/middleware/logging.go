package middleware

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		req := fmt.Sprintf("%s %s", r.Method, r.URL.Path)

		log.Printf(req)
		next.ServeHTTP(w, r)
		log.Printf("completed %s in %v", req, time.Since(start))
	})
}
