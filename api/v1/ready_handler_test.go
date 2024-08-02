package v1_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	json "github.com/json-iterator/go"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/jakobmoellerdev/octi-sync-server/api/v1/REST"
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

	res := REST.HealthAggregation{}

	assert.NoError(json.NewDecoder(rec.Body).Decode(&res))
	assert.NotEmpty(res.Health)
	assert.Equal(REST.Up, res.Health)

	assert.NotNil(res.Components)
	assert.NotEmpty(res.Components)
	assert.Len(*res.Components, 3)

	for _, component := range *res.Components {
		assert.Equal(REST.Up, component.Health)
		assert.NotEmpty(component.Name)
	}
}
