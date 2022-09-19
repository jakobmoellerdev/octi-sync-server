package request

import (
	"time"

	"github.com/didip/tollbooth/v7"
	"github.com/didip/tollbooth/v7/limiter"
	"github.com/gin-gonic/gin"
)

const (
	DefaultLimitMaxCallAmount        = 10
	DefaultLimitExpirationTTLSeconds = 1
)

func DefaultLimit() *limiter.Limiter {
	return tollbooth.NewLimiter(DefaultLimitMaxCallAmount, &limiter.ExpirableOptions{
		DefaultExpirationTTL: DefaultLimitExpirationTTLSeconds * time.Second,
	})
}

func LimitHandler(lmt *limiter.Limiter) gin.HandlerFunc {
	return func(context *gin.Context) {
		httpError := tollbooth.LimitByRequest(lmt, context.Writer, context.Request)
		if httpError != nil {
			context.Data(httpError.StatusCode, lmt.GetMessageContentType(), []byte(httpError.Message))
			context.Abort()
		} else {
			context.Next()
		}
	}
}
