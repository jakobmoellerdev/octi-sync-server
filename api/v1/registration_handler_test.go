package v1_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/uuid"
	json "github.com/json-iterator/go"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/jakob-moeller-cloud/octi-sync-server/api/v1/REST"
)

func TestAPIRegister(t *testing.T) {
	t.Parallel()
	_, assertions, api := SetupAPITest(t)

	req := httptest.NewRequest(http.MethodPost, "/",
		strings.NewReader(make(url.Values).Encode()))

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	deviceID, err := uuid.NewRandom()
	assertions.NoError(err)

	if rec := httptest.NewRecorder(); assertions.NoError(
		API().Register(api.NewContext(req, rec), REST.RegisterParams{XDeviceID: deviceID}),
	) {
		verifyRegistrationResponse(assertions, rec)
	}
}

func verifyRegistrationResponse(assert *assert.Assertions, rec *httptest.ResponseRecorder) {
	assert.Equal(http.StatusOK, rec.Code)

	res := REST.RegistrationResponse{}

	assert.NoError(json.NewDecoder(rec.Body).Decode(&res))
	assert.NotEmpty(res.DeviceID)
	assert.NotEmpty(res.Username)
	assert.NotEmpty(res.Password)
}
