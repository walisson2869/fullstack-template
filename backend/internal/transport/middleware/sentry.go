package middleware

import (
	sentry "github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"

	"github.com/gin-gonic/gin"
)

// SentryMiddleware reports panics and errors to Sentry.
// When dsn is empty it returns a no-op handler so the app works without Sentry configured.
func SentryMiddleware(dsn string) gin.HandlerFunc {
	if dsn == "" {
		return func(c *gin.Context) { c.Next() }
	}
	_ = sentry.Init(sentry.ClientOptions{Dsn: dsn})
	return sentrygin.New(sentrygin.Options{Repanic: true})
}
