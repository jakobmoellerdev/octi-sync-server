package logging_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	json "github.com/json-iterator/go"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"github.com/jakobmoellerdev/octi-sync-server/middleware/basic"
	"github.com/jakobmoellerdev/octi-sync-server/middleware/logging"
)

func TestRequestLogging(t *testing.T) {
	t.Parallel()

	logBuf := bytes.NewBufferString("")
	log := zerolog.New(zerolog.New(logBuf))
	rec := httptest.NewRecorder()
	req := emptyRequest(http.MethodGet)

	req.Header.Set("x-request-id", "test")
	req.Header.Set(basic.DeviceIDHeader, "test")

	ctx := echo.New().NewContext(req, rec)
	assertions := assert.New(t)

	err := logging.RequestLogging(&log)(func(ctx echo.Context) error {
		return ctx.String(http.StatusOK, "test") //nolint:wrapcheck
	})(ctx)

	assertions.NoError(err)

	logData := map[string]string{}

	assertions.NoError(json.NewDecoder(logBuf).Decode(&logData))

	assertions.NotEmpty(logData)
	assertions.Contains(logData, "message")

	message := map[string]any{}
	assertions.NoError(json.Unmarshal([]byte(logData["message"]), &message))
	assertions.Equal("/", message["URI"])
	assertions.Equal(float64(200), message["status"])
	assertions.Empty(message["content-length"])
	assertions.Equal("test", message["x-request-id"])
	assertions.Equal("test", message["x-device-id"])
	assertions.Equal(float64(4), message["response-size"])
	assertions.NotNil(message["latency"])
	assertions.Equal("request", message["message"])
	assertions.Equal(zerolog.LevelDebugValue, message["level"])
	assertions.Equal(http.MethodGet, message["Method"])
	assertions.Empty(message["user-agent"])
	assertions.NotEmpty(message["remote-ip"])
}

func emptyRequest(method string) *http.Request {
	return httptest.NewRequest(method, "/",
		strings.NewReader(make(url.Values).Encode()))
}
