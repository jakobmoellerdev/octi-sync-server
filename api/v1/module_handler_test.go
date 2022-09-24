package v1_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/jakob-moeller-cloud/octi-sync-server/api/v1/REST"
	"github.com/jakob-moeller-cloud/octi-sync-server/service/memory"
	json "github.com/json-iterator/go"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestAPI_CreateModule(t *testing.T) {
	t.Parallel()
	_, assertions, api := SetupAPITest(t)

	req := emptyRequest(http.MethodPost)

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	deviceID, err := uuid.NewRandom()
	assertions.NoError(err)

	if rec := httptest.NewRecorder(); assertions.NoError(
		API().CreateModule(api.NewContext(req, rec), "test", REST.CreateModuleParams{XDeviceID: deviceID}),
	) {
		verifyCreateModuleResponse(assertions, rec)
	}
}

func verifyCreateModuleResponse(assert *assert.Assertions, rec *httptest.ResponseRecorder) {
	assert.Equal(http.StatusAccepted, rec.Code)
	assert.NoError(json.Unmarshal(rec.Body.Bytes(), &map[string]string{}))
}

func TestAPI_GetModule(t *testing.T) {
	t.Parallel()
	_, assertions, api := SetupAPITest(t)

	req := emptyRequest(http.MethodPost)

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	moduleName := "test"
	deviceID, err := uuid.NewRandom()
	assertions.NoError(err)

	apiImpl := API()

	assertions.NoError(apiImpl.Modules.Set(
		context.Background(), fmt.Sprintf("%s-%s", deviceID.String(), moduleName),
		memory.ModuleFromBytes([]byte("test"))),
	)

	if rec := httptest.NewRecorder(); assertions.NoError(
		apiImpl.GetModule(api.NewContext(req, rec), moduleName, REST.GetModuleParams{XDeviceID: deviceID}),
	) {
		verifyGetModuleResponse(assertions, rec)
	}
}

func verifyGetModuleResponse(assert *assert.Assertions, rec *httptest.ResponseRecorder) {
	assert.Equal(http.StatusOK, rec.Code)

	body := rec.Body.String()
	assert.NotEmpty(body)

	assert.Equal("test", body)
}
