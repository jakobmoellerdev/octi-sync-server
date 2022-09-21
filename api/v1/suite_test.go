package v1_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	v1 "github.com/jakob-moeller-cloud/octi-sync-server/api/v1"
	"github.com/jakob-moeller-cloud/octi-sync-server/middleware/logging"
	"github.com/jakob-moeller-cloud/octi-sync-server/service/memory"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func SetupAPITest(t *testing.T) (*zerolog.Logger, *assert.Assertions, *echo.Echo) {
	t.Helper()
	logger := zerolog.New(zerolog.NewTestWriter(t))
	api := echo.New()
	api.Use(logging.RequestLogging(&logger))

	return &logger, assert.New(t), api
}

func API() *v1.API {
	return &v1.API{
		Accounts: memory.NewAccounts(),
		Devices:  memory.NewDevices(),
		Modules:  memory.NewModules(),
	}
}

func emptyRequest(method string) *http.Request {
	return httptest.NewRequest(method, "/",
		strings.NewReader(make(url.Values).Encode()))
}

func testAPI(
	t *testing.T,
	api *echo.Echo,
	assertions *assert.Assertions,
	request *http.Request,
	handler echo.HandlerFunc,
	verify func(*assert.Assertions, *httptest.ResponseRecorder),
) {
	t.Helper()

	if rec := httptest.NewRecorder(); assertions.NoError(
		handler(api.NewContext(request, rec)),
		"Call to Handler should not fail",
	) {
		verify(assertions, rec)
	}
}
