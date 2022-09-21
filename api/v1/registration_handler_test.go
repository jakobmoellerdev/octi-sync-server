package v1_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	v1 "github.com/jakob-moeller-cloud/octi-sync-server/api/v1"
	"github.com/jakob-moeller-cloud/octi-sync-server/middleware/logging"
	"github.com/jakob-moeller-cloud/octi-sync-server/service/mem"
	jsoniter "github.com/json-iterator/go"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestAPIRegister(t *testing.T) {
	t.Parallel()

	_, assertions, api := SetupAPITest(t)

	req := httptest.NewRequest(http.MethodPost, "/",
		strings.NewReader(make(url.Values).Encode()))

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	if rec := httptest.NewRecorder(); assertions.NoError(
		v1API().Register(api.NewContext(req, rec), v1.RegisterParams{XDeviceID: "test"}),
	) {
		verifyRegistrationResponse(assertions, rec)
	}
}

func SetupAPITest(t *testing.T) (*zerolog.Logger, *assert.Assertions, *echo.Echo) {
	t.Helper()
	logger := zerolog.New(zerolog.NewTestWriter(t))
	api := echo.New()
	api.Use(logging.RequestLogging(&logger))

	return &logger, assert.New(t), api
}

func v1API() *v1.API {
	return &v1.API{
		Accounts: mem.NewAccounts(),
		Devices:  mem.NewDevices(),
	}
}

func verifyRegistrationResponse(assertions *assert.Assertions, rec *httptest.ResponseRecorder) {
	assertions.Equal(http.StatusOK, rec.Code)

	res := v1.RegistrationResponse{}

	assertions.NoError(jsoniter.NewDecoder(rec.Body).Decode(&res))
	assertions.NotEmpty(res.DeviceID)
	assertions.NotEmpty(res.Username)
	assertions.NotEmpty(res.Password)
}
