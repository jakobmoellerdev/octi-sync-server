package v1_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	v1 "github.com/jakob-moeller-cloud/octi-sync-server/api/v1"
	json "github.com/json-iterator/go"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestAPI_IsReady(t *testing.T) {
	t.Parallel()
	_, assertions, api := SetupAPITest(t)

	req := emptyRequest(http.MethodGet)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	testAPI(t, api, assertions, req, API().IsReady, verifyIsReady)
}

func verifyIsReady(assert *assert.Assertions, rec *httptest.ResponseRecorder) {
	assert.Equal(http.StatusOK, rec.Code)

	res := v1.HealthAggregation{}

	assert.NoError(json.NewDecoder(rec.Body).Decode(&res))
	assert.NotEmpty(res.Health)
	assert.Equal(v1.Up, res.Health)

	assert.NotNil(res.Components)
	assert.NotEmpty(res.Components)
	assert.Len(*res.Components, 3)

	for _, component := range *res.Components {
		assert.Equal(v1.Up, component.Health)
		assert.NotEmpty(component.Name)
	}
}
