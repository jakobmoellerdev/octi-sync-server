package v1_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/jakob-moeller-cloud/octi-sync-server/api/v1/REST"
	"github.com/jakob-moeller-cloud/octi-sync-server/middleware/basic"
	"github.com/jakob-moeller-cloud/octi-sync-server/service/memory"
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

	deviceID, err := uuid.NewRandom()
	assertions.NoError(err)

	assertions.Error(echo.ErrForbidden,
		apiImpl.Share(api.NewContext(req, rec), REST.ShareParams{XDeviceID: deviceID}))

	user := "test-user"
	ctx := api.NewContext(req, rec)

	ctx.Set(basic.AccountKey, memory.NewAccount(user, ""))

	if assertions.NoError(apiImpl.Share(ctx, REST.ShareParams{XDeviceID: deviceID})) {
		verifyShare(assertions, rec)
	}
}

func verifyShare(assert *assert.Assertions, rec *httptest.ResponseRecorder) {
	assert.Equal(http.StatusOK, rec.Code)

	res := REST.ShareResponse{}

	assert.NoError(json.NewDecoder(rec.Body).Decode(&res))
	assert.NotEmpty(res.ShareCode)
}
