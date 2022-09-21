package v1_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	v1 "github.com/jakob-moeller-cloud/octi-sync-server/api/v1"
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

	if rec := httptest.NewRecorder(); assertions.NoError(
		API().CreateModule(api.NewContext(req, rec), "test", v1.CreateModuleParams{XDeviceID: "test"}),
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
	deviceName := "test"

	apiImpl := API()

	assertions.NoError(apiImpl.Modules.Set(
		context.Background(), fmt.Sprintf("%s-%s", deviceName, moduleName),
		memory.ModuleFromBytes([]byte("test"))),
	)

	if rec := httptest.NewRecorder(); assertions.NoError(
		apiImpl.GetModule(api.NewContext(req, rec), moduleName, v1.GetModuleParams{XDeviceID: deviceName}),
	) {
		verifyGetModuleResponse(assertions, rec)
	}
}

func verifyGetModuleResponse(assert *assert.Assertions, rec *httptest.ResponseRecorder) {
	assert.Equal(http.StatusOK, rec.Code)

	body := string(rec.Body.Bytes())
	assert.NotEmpty(body)

	assert.Equal("test", body)
}
