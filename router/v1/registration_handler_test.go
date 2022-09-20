package v1_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/jakob-moeller-cloud/octi-sync-server/middleware/logging"
	v1 "github.com/jakob-moeller-cloud/octi-sync-server/router/v1"
	"github.com/jakob-moeller-cloud/octi-sync-server/service/mem"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestAPIRegister(t *testing.T) {
	t.Parallel()
	_, assertions, api := SetupAPITest(t)

	req := httptest.NewRequest(http.MethodPost, "/",
		strings.NewReader(make(url.Values).Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := api.NewContext(req, rec)

	if assertions.NoError(memoryRegistration().Register(c)) {
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

func memoryRegistration() *v1.RegistrationHandler {
	return &v1.RegistrationHandler{
		Accounts: mem.NewAccounts(),
		Devices:  mem.NewDevices(),
	}
}

func verifyRegistrationResponse(assertions *assert.Assertions, rec *httptest.ResponseRecorder) {
	assertions.Equal(http.StatusOK, rec.Code)
	res := v1.RegistrationResponse{}
	assertions.NoError(yaml.NewDecoder(rec.Body).Decode(&res))
	assertions.NotEmpty(res.DeviceID)
	assertions.NotEmpty(res.Username)
	assertions.NotEmpty(res.Password)
}
