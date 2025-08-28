package middlewares

import (
	"fmt"
	"net/http"
)

var allowedOrigin = []string{
	"https://localhost:3000",
	"http://localhost:3000",
	"http://localhost:5173",
	"https://localhost:5173",
	"http://127.0.0.1:5173",
	"",
}

func Cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("cors middleware begins...")

		origin := r.Header.Get("Origin")
		fmt.Println("origin is :", origin)

		if isOriginAllowed(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			http.Error(w, "Origin Not Allowed...", http.StatusForbidden)
			return
		}

		// Set other CORS headers
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Expose-Headers", "Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// Handle preflight request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
		fmt.Println("cors middleware ends...")

	})
}

func isOriginAllowed(origin string) bool {
	for _, value := range allowedOrigin {
		if origin == value {
			fmt.Println("origin verified....")
			return true
		}
	}
	return false
}
