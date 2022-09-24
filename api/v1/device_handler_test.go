package v1_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	v1 "github.com/jakob-moeller-cloud/octi-sync-server/api/v1"
	"github.com/jakob-moeller-cloud/octi-sync-server/api/v1/REST"
	"github.com/jakob-moeller-cloud/octi-sync-server/middleware/basic"
	"github.com/jakob-moeller-cloud/octi-sync-server/service"
	"github.com/jakob-moeller-cloud/octi-sync-server/service/mock"
	json "github.com/json-iterator/go"
	"github.com/labstack/echo/v4"
	assertions "github.com/stretchr/testify/assert"
)

type deviceTest func(ctx echo.Context, rec *httptest.ResponseRecorder)

func TestAPI_GetDevices(t *testing.T) {
	t.Parallel()

	assert, ctrl := assertions.New(t), gomock.NewController(t)
	router := echo.New()
	devices := mock.NewMockDevices(ctrl)

	api := &v1.API{
		Devices: devices,
	}

	for _, testCase := range []struct {
		name string
		test deviceTest
	}{
		{
			"fail forbidden as not account is present",
			func(ctx echo.Context, rec *httptest.ResponseRecorder) {
				err := api.GetDevices(ctx, REST.GetDevicesParams{XDeviceID: RandomUUID(t)})
				assert.ErrorIs(err, v1.ErrNoDeviceAccessWithoutAccount)
			},
		},
		{
			"failure during account fetching returns 500",
			testGetDevicesReturns500(t, api),
		},
		{
			"succeed as account is present",
			testGetDevicesReturns200(t, api),
		},
	} {
		req := emptyRequest(http.MethodGet)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		ctx := router.NewContext(req, rec)

		testCase.test(ctx, rec)

		if t.Failed() {
			break
		}
	}
}

func mockDevicesFromAPI(t *testing.T, api *v1.API) *mock.MockDevices {
	devices, isMock := api.Devices.(*mock.MockDevices)
	if !isMock {
		assertions.New(t).FailNow("devices are not mocked")
	}

	return devices
}

func testGetDevicesReturns200(t *testing.T, api *v1.API) deviceTest {
	assert := assertions.New(t)

	return func(ctx echo.Context, rec *httptest.ResponseRecorder) {
		acc := service.NewBaseAccount("test", "pw")
		ctx.Set(basic.AccountKey, acc)

		deviceID := service.DeviceID(RandomUUID(t))
		mockDevicesFromAPI(t, api).EXPECT().FindByAccount(context.Background(), acc).Times(1).
			Return([]service.Device{
				service.DeviceFromID(deviceID),
			}, nil)

		err := api.GetDevices(ctx, REST.GetDevicesParams{XDeviceID: REST.XDeviceID(deviceID)})
		assert.NoError(err)
		assert.Equal(http.StatusOK, rec.Code)
		assert.NotNil(rec.Body)
		assert.NotEmpty(rec.Body)

		var deviceListResponse REST.DeviceListResponse

		assert.NoError(json.Unmarshal(rec.Body.Bytes(), &deviceListResponse))
		assert.Len(deviceListResponse.Items, deviceListResponse.Count,
			"list count should equal item count")
	}
}

func testGetDevicesReturns500(t *testing.T, api *v1.API) deviceTest {
	assert := assertions.New(t)

	return func(ctx echo.Context, rec *httptest.ResponseRecorder) {
		acc := service.NewBaseAccount("test", "pw")
		ctx.Set(basic.AccountKey, acc)

		mockDevicesFromAPI(t, api).EXPECT().FindByAccount(context.Background(), acc).Times(1).
			Return(nil, errors.New("mock account err"))

		err := api.GetDevices(ctx, REST.GetDevicesParams{XDeviceID: RandomUUID(t)})
		httpError, isHTTPError := err.(*echo.HTTPError)

		assert.True(isHTTPError)
		assert.Equal(http.StatusInternalServerError, httpError.Code)

		messageErr, messageIsErr := httpError.Message.(error)

		assert.True(messageIsErr)
		assert.ErrorContains(messageErr, "could not fetch devices from account")
	}
}
