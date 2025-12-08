package main

import (
	"log"
	"net/http"
)

func (s *server) authRequired(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("auth_token")
		if err != nil {
			http.Error(w, "no autenticado", http.StatusUnauthorized)
			return
		}

		claims, err := s.auth.validateJWT(cookie.Value)
		if err != nil {
			http.Error(w, "token inválido o expirado", http.StatusUnauthorized)
			return
		}

		if s.auth.requiredGuildID != "" && claims.GuildID != s.auth.requiredGuildID {
			http.Error(w, "debes ser miembro del servidor de Discord para usar Wasabi", http.StatusForbidden)
			return
		}

		ctx := contextWithUser(r.Context(), claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func corsMiddleware(allowedOrigins []string, next http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		if origin == "" {
			continue
		}
		allowed[origin] = struct{}{}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			if _, ok := allowed[origin]; ok {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.Header().Add("Vary", "Origin")
			}
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
