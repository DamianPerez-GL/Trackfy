package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/trackfy/api-gateway/internal/db"
)

type RateLimiter struct {
	redis *db.RedisDB
}

func NewRateLimiter(redis *db.RedisDB) *RateLimiter {
	return &RateLimiter{redis: redis}
}

// Limit crea un middleware de rate limiting
func (rl *RateLimiter) Limit(requests int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Usar IP o UserID como clave
			key := getClientIdentifier(r)

			allowed, count, err := rl.redis.CheckRateLimit(r.Context(), key, requests, window)
			if err != nil {
				log.Error().Err(err).Msg("[RateLimit] Redis error")
				// En caso de error, permitir la request
				next.ServeHTTP(w, r)
				return
			}

			// Añadir headers de rate limit
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", requests))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", max(0, requests-count)))
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(window).Unix()))

			if !allowed {
				w.Header().Set("Retry-After", fmt.Sprintf("%d", int(window.Seconds())))
				respondError(w, http.StatusTooManyRequests, "rate_limit_exceeded", "Too many requests")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// LimitByUser rate limit por usuario autenticado
func (rl *RateLimiter) LimitByUser(requests int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := GetUserID(r.Context())
			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			key := "user:" + userID.String()

			allowed, count, err := rl.redis.CheckRateLimit(r.Context(), key, requests, window)
			if err != nil {
				log.Error().Err(err).Msg("[RateLimit] Redis error")
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", requests))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", max(0, requests-count)))

			if !allowed {
				respondError(w, http.StatusTooManyRequests, "rate_limit_exceeded", "Too many requests")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func getClientIdentifier(r *http.Request) string {
	// Intentar obtener UserID del contexto
	if userID, ok := GetUserID(r.Context()); ok {
		return "user:" + userID.String()
	}

	// Usar IP como fallback
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}
	if ip == "" {
		ip = r.RemoteAddr
	}

	return "ip:" + ip
}
