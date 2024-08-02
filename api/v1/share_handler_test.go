package v1_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	json "github.com/json-iterator/go"
	"github.com/labstack/echo/v4"
	assertions "github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	v1 "github.com/jakobmoellerdev/octi-sync-server/api/v1"
	"github.com/jakobmoellerdev/octi-sync-server/api/v1/REST"
	"github.com/jakobmoellerdev/octi-sync-server/middleware/basic"
	"github.com/jakobmoellerdev/octi-sync-server/service"
	"github.com/jakobmoellerdev/octi-sync-server/service/mock"
)

func TestAPI_Share(t *testing.T) {
	t.Parallel()
	assert := assertions.New(t)
	api := echo.New()
	ctrl := gomock.NewController(t)
	sharing := mock.NewMockSharing(ctrl)
	apiImpl := &v1.API{
		Sharing: sharing,
	}
	rec := httptest.NewRecorder()
	user := "test-user"

	req := emptyRequest(http.MethodGet)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	deviceID, err := uuid.NewRandom()
	assert.NoError(err)

	ctx := api.NewContext(req, rec)

	assert.Error(
		echo.ErrForbidden,
		apiImpl.Share(api.NewContext(req, rec), REST.ShareParams{XDeviceID: deviceID}),
	)

	account := service.NewBaseAccount(user, time.Now())
	ctx.Set(basic.AccountKey, account)

	shareCode := "share"

	sharing.EXPECT().Share(context.Background(), gomock.Any()).Times(1).Return(service.ShareCode(shareCode), nil)

	if assert.NoError(apiImpl.Share(ctx, REST.ShareParams{XDeviceID: deviceID})) {
		assert.Equal(http.StatusOK, rec.Code)

		res := REST.ShareResponse{}

		assert.NoError(json.NewDecoder(rec.Body).Decode(&res))
		assert.Equal(shareCode, *res.ShareCode)
	}
}
