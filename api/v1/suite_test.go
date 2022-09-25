package v1_test

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	v1 "github.com/jakob-moeller-cloud/octi-sync-server/api/v1"
	"github.com/jakob-moeller-cloud/octi-sync-server/middleware/logging"
	"github.com/jakob-moeller-cloud/octi-sync-server/service/memory"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

// Lock to avoid race when creating Log Writers.
var apiSetup = sync.Mutex{} //nolint:gochecknoglobals

func SetupAPITest(t *testing.T) (zerolog.Logger, *assert.Assertions, *echo.Echo) {
	apiSetup.Lock()
	defer apiSetup.Unlock()
	t.Helper()
	logger := zerolog.New(zerolog.NewConsoleWriter(zerolog.ConsoleTestWriter(t)))
	api := echo.New()
	api.Use(logging.RequestLogging(&logger))

	return logger, assert.New(t), api
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

func RandomUUID(t *testing.T) uuid.UUID {
	t.Helper()

	newUUID, err := uuid.NewRandom()

	assert.NoError(t, err)

	return newUUID
}

func HashedPassword(password string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(password)))
}
