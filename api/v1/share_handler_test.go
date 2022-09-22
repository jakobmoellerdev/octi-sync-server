package v1_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	v1 "github.com/jakob-moeller-cloud/octi-sync-server/api/v1"
	"github.com/jakob-moeller-cloud/octi-sync-server/middleware/auth"
	json "github.com/json-iterator/go"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestAPI_Share(t *testing.T) {
	t.Parallel()
	_, assertions, api := SetupAPITest(t)

	req := emptyRequest(http.MethodGet)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	apiImpl := API()
	rec := httptest.NewRecorder()

	assertions.Error(echo.ErrForbidden,
		apiImpl.Share(api.NewContext(req, rec), v1.ShareParams{XDeviceID: "test"}))

	user := "test-user"
	ctx := api.NewContext(req, rec)

	ctx.Set(auth.UserKey, user)

	if assertions.NoError(apiImpl.Share(ctx, v1.ShareParams{XDeviceID: "test"})) {
		verifyShare(assertions, rec)
	}
}

func verifyShare(assert *assert.Assertions, rec *httptest.ResponseRecorder) {
	assert.Equal(http.StatusOK, rec.Code)

	res := v1.ShareResponse{}

	assert.NoError(json.NewDecoder(rec.Body).Decode(&res))
	assert.NotEmpty(res.ShareCode)
}
