package middleware

import (
	"time"

	"github.com/didip/tollbooth/v7"
	"github.com/didip/tollbooth/v7/limiter"
	"github.com/saikey0379/go-json-rest/rest"
	"golang.org/x/net/context"

	"github.com/saikey0379/imp-server/pkg/logger"
)

// LimiterMiddleware cancel context when timeout
type LimiterMiddleware struct {
	logger  logger.Logger
	limiter *limiter.Limiter
}

// Limiter create timeout middleware with duration
func Limiter(logger logger.Logger, limiter *limiter.Limiter) *LimiterMiddleware {
	return &LimiterMiddleware{logger, limiter}
}

// NewLimiterMiddleware create a simple limiter
func NewLimiterMiddleware(logger logger.Logger, max float64, ttl time.Duration) *LimiterMiddleware {
	return Limiter(logger, tollbooth.NewLimiter(max, &limiter.ExpirableOptions{DefaultExpirationTTL: ttl}))
}

// MiddlewareFunc makes LimiterMiddleware implement the Middleware interface.
func (mw *LimiterMiddleware) MiddlewareFunc(h rest.HandlerFunc) rest.HandlerFunc {
	return func(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
		httpError := tollbooth.LimitByRequest(mw.limiter, w, r.Request)
		if httpError != nil {
			mw.logger.Warnf("%s\n", httpError.Message)
			w.WriteHeader(httpError.StatusCode)
			w.WriteJSON(map[string]string{"status": "error", "msg": httpError.Message})
			return
		}

		// There's no rate-limit error, serve the next handler.
		h(ctx, w, r)
	}
}
