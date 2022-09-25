package v1_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	v1 "github.com/jakob-moeller-cloud/octi-sync-server/api/v1"
	"github.com/jakob-moeller-cloud/octi-sync-server/api/v1/REST"
	"github.com/jakob-moeller-cloud/octi-sync-server/service"
	"github.com/jakob-moeller-cloud/octi-sync-server/service/mock"
	"github.com/labstack/echo/v4"
	"github.com/sethvargo/go-password/password"
	assertions "github.com/stretchr/testify/assert"
)

type registerTest func(assert *assertions.Assertions, ctx echo.Context, rec *httptest.ResponseRecorder)

func TestAPIRegister(t *testing.T) {
	t.Parallel()

	router := echo.New()

	errMockText := "mocked error"
	user := "test-user"
	pass := "test-pass"
	deviceID := service.DeviceID(uuid.Must(uuid.NewRandom()))
	share := "test"

	var ctrl *gomock.Controller

	var devices *mock.MockDevices

	var accounts *mock.MockAccounts

	var api *v1.API

	for _, testCase := range []struct {
		name string
		test registerTest
	}{
		{
			"500 as credential/username generation during registration fails",
			func(assert *assertions.Assertions, ctx echo.Context, rec *httptest.ResponseRecorder) {
				usernameGen := mock.NewMockUsernameGenerator(ctrl)
				usernameGen.EXPECT().Generate().AnyTimes().Return("", errors.New(errMockText))
				api.UsernameGenerator = usernameGen
				err := api.Register(ctx, REST.RegisterParams{})
				assert.ErrorContains(err, "generating a username for registration failed")
				assert.ErrorContains(err, errMockText)
			},
		},
		{
			"500 as credential/password generation during registration fails",
			func(assert *assertions.Assertions, ctx echo.Context, rec *httptest.ResponseRecorder) {
				api.PasswordGenerator = password.NewMockGenerator("", errors.New(errMockText))
				err := api.Register(ctx, REST.RegisterParams{})
				assert.ErrorContains(err, "generating a password for registration failed")
				assert.ErrorContains(err, errMockText)
			},
		},
		{
			"500 as account registration fails",
			func(assert *assertions.Assertions, ctx echo.Context, rec *httptest.ResponseRecorder) {
				accounts.EXPECT().Register(ctx.Request().Context(), user, pass).Times(1).
					Return(nil, errors.New(errMockText))
				err := api.Register(ctx, REST.RegisterParams{})
				assert.ErrorContains(err, errMockText)
			},
		},
		{
			"403 as account does not exist but credentials are provided",
			func(assert *assertions.Assertions, ctx echo.Context, rec *httptest.ResponseRecorder) {
				accounts.EXPECT().Find(ctx.Request().Context(), "foo").Times(1).
					Return(nil, errors.New(errMockText))
				ctx.Request().SetBasicAuth("foo", "bar")
				err := api.Register(ctx, REST.RegisterParams{})
				assert.ErrorContains(err, errMockText)
			},
		},
		{
			"403 as account does exist but pass is wrong",
			func(assert *assertions.Assertions, ctx echo.Context, rec *httptest.ResponseRecorder) {
				accounts.EXPECT().Find(ctx.Request().Context(), "foo").Times(1).
					Return(service.NewBaseAccount("foo", "bar"), nil)
				ctx.Request().SetBasicAuth("foo", "WRONG")
				err := api.Register(ctx, REST.RegisterParams{})
				assert.ErrorIs(err, v1.ErrPasswordMismatch)
			},
		},
		{
			"403 as device is not in account and no share code is provided",
			func(assert *assertions.Assertions, ctx echo.Context, rec *httptest.ResponseRecorder) {
				acc := service.NewBaseAccount(user, HashedPassword(pass))
				ctx.Request().SetBasicAuth(user, pass)
				accounts.EXPECT().Find(ctx.Request().Context(), user).Times(1).
					Return(acc, nil)
				devices.EXPECT().FindByDeviceID(ctx.Request().Context(), acc, deviceID).Times(1).
					Return(nil, service.ErrDeviceNotFound)
				err := api.Register(ctx, REST.RegisterParams{
					XDeviceID: REST.XDeviceID(deviceID),
				})
				assert.ErrorIs(err, service.ErrDeviceNotFound)
			},
		},
		{
			"403 as device is not in account and share code is provided but share lookup fails",
			func(assert *assertions.Assertions, ctx echo.Context, rec *httptest.ResponseRecorder) {
				acc := service.NewBaseAccount(user, HashedPassword(pass))
				ctx.Request().SetBasicAuth(user, pass)
				accounts.EXPECT().Find(ctx.Request().Context(), user).Times(1).
					Return(acc, nil)
				devices.EXPECT().FindByDeviceID(ctx.Request().Context(), acc, deviceID).Times(1).
					Return(nil, service.ErrDeviceNotFound)
				accounts.EXPECT().IsShared(ctx.Request().Context(), acc.Username(), share).Times(1).
					Return(errors.New(errMockText))
				err := api.Register(ctx, REST.RegisterParams{
					XDeviceID: REST.XDeviceID(deviceID),
					Share:     &share,
				})
				assert.ErrorContains(err, "cannot verify share")
			},
		},
		{
			"403 as device is not in account and share code is provided but share is invalid",
			func(assert *assertions.Assertions, ctx echo.Context, rec *httptest.ResponseRecorder) {
				acc := service.NewBaseAccount(user, HashedPassword(pass))
				ctx.Request().SetBasicAuth(user, pass)
				accounts.EXPECT().Find(ctx.Request().Context(), user).Times(1).
					Return(acc, nil)
				devices.EXPECT().FindByDeviceID(ctx.Request().Context(), acc, deviceID).Times(1).
					Return(nil, service.ErrDeviceNotFound)
				accounts.EXPECT().IsShared(ctx.Request().Context(), acc.Username(), share).Times(1).
					Return(service.ErrShareCodeInvalid)
				err := api.Register(ctx, REST.RegisterParams{
					XDeviceID: REST.XDeviceID(deviceID),
					Share:     &share,
				})
				assert.ErrorContains(err, "is invalid (not shared)")
			},
		},
		{
			"500 as device is not in account and share code is provided and valid but device registration fails",
			func(assert *assertions.Assertions, ctx echo.Context, rec *httptest.ResponseRecorder) {
				acc := service.NewBaseAccount(user, HashedPassword(pass))
				ctx.Request().SetBasicAuth(user, pass)
				accounts.EXPECT().Find(ctx.Request().Context(), user).Times(1).
					Return(acc, nil)
				devices.EXPECT().FindByDeviceID(ctx.Request().Context(), acc, deviceID).Times(1).
					Return(nil, service.ErrDeviceNotFound)
				accounts.EXPECT().IsShared(ctx.Request().Context(), acc.Username(), share).Times(1).
					Return(nil)
				devices.EXPECT().Register(ctx.Request().Context(), acc, deviceID).Times(1).
					Return(nil, errors.New(errMockText))
				err := api.Register(ctx, REST.RegisterParams{
					XDeviceID: REST.XDeviceID(deviceID),
					Share:     &share,
				})
				assert.ErrorContains(err, "cannot register device")
			},
		},
		{
			"200 as device is not in account and share code is provided and valid and device registration " +
				"succeeds",
			func(assert *assertions.Assertions, ctx echo.Context, rec *httptest.ResponseRecorder) {
				acc := service.NewBaseAccount(user, HashedPassword(pass))
				ctx.Request().SetBasicAuth(user, pass)
				accounts.EXPECT().Find(ctx.Request().Context(), user).Times(1).
					Return(acc, nil)
				devices.EXPECT().FindByDeviceID(ctx.Request().Context(), acc, deviceID).Times(1).
					Return(nil, service.ErrDeviceNotFound)
				accounts.EXPECT().IsShared(ctx.Request().Context(), acc.Username(), share).Times(1).
					Return(nil)
				devices.EXPECT().Register(ctx.Request().Context(), acc, deviceID).Times(1).
					Return(service.NewBaseDevice(deviceID), nil)
				err := api.Register(ctx, REST.RegisterParams{
					XDeviceID: REST.XDeviceID(deviceID),
					Share:     &share,
				})
				assert.NoError(err)
			},
		},
		{
			"200 as device is not in account and share code is provided and valid and device registration " +
				"succeeds with basic auth header",
			func(assert *assertions.Assertions, ctx echo.Context, rec *httptest.ResponseRecorder) {
				acc := service.NewBaseAccount(user, HashedPassword(pass))
				ctx.Request().SetBasicAuth(user, pass)
				accounts.EXPECT().Find(ctx.Request().Context(), user).Times(1).
					Return(acc, nil)
				devices.EXPECT().FindByDeviceID(ctx.Request().Context(), acc, deviceID).Times(1).
					Return(nil, service.ErrDeviceNotFound)
				accounts.EXPECT().IsShared(ctx.Request().Context(), acc.Username(), share).Times(1).
					Return(nil)
				devices.EXPECT().Register(ctx.Request().Context(), acc, deviceID).Times(1).
					Return(service.NewBaseDevice(deviceID), nil)
				err := api.Register(ctx, REST.RegisterParams{
					XDeviceID: REST.XDeviceID(deviceID),
					Share:     &share,
				})
				assert.NoError(err)
			},
		},
		{
			"200 as device is not in account and share code is provided and valid and device registration " +
				"succeeds with generated credentials",
			func(assert *assertions.Assertions, ctx echo.Context, rec *httptest.ResponseRecorder) {
				acc := service.NewBaseAccount(user, HashedPassword(pass))
				accounts.EXPECT().Register(ctx.Request().Context(), user, pass).Times(1).
					Return(acc, nil)
				devices.EXPECT().FindByDeviceID(ctx.Request().Context(), acc, deviceID).Times(1).
					Return(nil, service.ErrDeviceNotFound)
				accounts.EXPECT().IsShared(ctx.Request().Context(), acc.Username(), share).Times(1).
					Return(nil)
				devices.EXPECT().Register(ctx.Request().Context(), acc, deviceID).Times(1).
					Return(service.NewBaseDevice(deviceID), nil)
				err := api.Register(ctx, REST.RegisterParams{
					XDeviceID: REST.XDeviceID(deviceID),
					Share:     &share,
				})
				assert.NoError(err)
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			ctrl = gomock.NewController(t)
			devices = mock.NewMockDevices(ctrl)
			accounts = mock.NewMockAccounts(ctrl)
			usernameGen := mock.NewMockUsernameGenerator(ctrl)
			usernameGen.EXPECT().Generate().AnyTimes().Return(user, nil)

			api = &v1.API{
				Devices:           devices,
				Accounts:          accounts,
				UsernameGenerator: usernameGen,
				PasswordGenerator: password.NewMockGenerator(pass, nil),
			}

			req := emptyRequest(http.MethodGet)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			rec := httptest.NewRecorder()
			ctx := router.NewContext(req, rec)
			testCase.test(assertions.New(t), ctx, rec)
		})
	}
}
