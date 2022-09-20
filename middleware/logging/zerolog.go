package logging

import (
	authmiddleware "github.com/jakob-moeller-cloud/octi-sync-server/middleware/auth"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"net"
)

func RequestLogging(logger *zerolog.Logger) echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:           true,
		LogStatus:        true,
		LogError:         true,
		LogLatency:       true,
		LogMethod:        true,
		LogContentLength: true,
		LogRemoteIP:      true,
		LogRequestID:     true,
		LogResponseSize:  true,
		LogUserAgent:     true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			logger.Debug().
				Str("Method", v.Method).
				Str("URI", v.URI).
				Int("status", v.Status).
				Str("content-length", v.ContentLength).
				Str("x-request-id", v.RequestID).
				Str("x-device-id", c.Request().Header.Get(authmiddleware.DeviceIDHeader)).
				Int64("response-size", v.ResponseSize).
				Str("user-agent", v.UserAgent).
				Err(v.Error).
				Dur("latency", v.Latency).
				IPAddr("remote-ip", net.ParseIP(v.RemoteIP)).
				Msg("request")
			return nil
		},
	})
}
