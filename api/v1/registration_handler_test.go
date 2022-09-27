package v1_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	v1 "github.com/jakob-moeller-cloud/octi-sync-server/api/v1"
	"github.com/jakob-moeller-cloud/octi-sync-server/api/v1/REST"
	"github.com/jakob-moeller-cloud/octi-sync-server/service"
	"github.com/jakob-moeller-cloud/octi-sync-server/service/mock"
	"github.com/labstack/echo/v4"
	"github.com/sethvargo/go-password/password"
	"github.com/stretchr/testify/suite"
)

func TestRegisterTestSuite(t *testing.T) {
	suite.Run(
		t, &RegisterTestSuite{},
	)
}

type RegisterTestSuite struct {
	suite.Suite
	ctrl     *gomock.Controller
	devices  *mock.MockDevices
	accounts *mock.MockAccounts
	sharing  *mock.MockSharing
	router   *echo.Echo
	api      *v1.API

	errMockText string
	user        string
	pass        string
	deviceID    service.DeviceID
	share       service.ShareCode

	rec *httptest.ResponseRecorder
	ctx echo.Context
	req *http.Request
}

func (r *RegisterTestSuite) SetupSuite() {
	r.errMockText = "mocked error"
	r.user = "test-user"
	r.pass = "test-pass"
	r.deviceID = service.DeviceID(uuid.Must(uuid.NewRandom()))
	r.share = "test"

	r.router = echo.New()
}

func (r *RegisterTestSuite) SetupTest() {
	r.ctrl = gomock.NewController(r.T())
	r.devices = mock.NewMockDevices(r.ctrl)
	r.accounts = mock.NewMockAccounts(r.ctrl)
	r.sharing = mock.NewMockSharing(r.ctrl)
	usernameGen := mock.NewMockUsernameGenerator(r.ctrl)
	usernameGen.EXPECT().Generate().AnyTimes().Return(r.user, nil)

	r.api = &v1.API{
		Devices:           r.devices,
		Accounts:          r.accounts,
		Sharing:           r.sharing,
		UsernameGenerator: usernameGen,
		PasswordGenerator: password.NewMockGenerator(r.pass, nil),
	}

	req := emptyRequest(http.MethodGet)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	r.req = req
	r.rec = httptest.NewRecorder()
	r.ctx = r.router.NewContext(req, r.rec)
}

func (r *RegisterTestSuite) Register(params REST.RegisterParams) error {
	return r.api.Register(r.ctx, params) //nolint:wrapcheck
}

func (r *RegisterTestSuite) Test_500_credential_username_generation_during_registration_fails() {
	usernameGen := mock.NewMockUsernameGenerator(r.ctrl)

	usernameGen.EXPECT().Generate().AnyTimes().Return("", errors.New(r.errMockText))

	r.api.UsernameGenerator = usernameGen
	err := r.Register(REST.RegisterParams{})

	r.ErrorContains(err, "generating a username for registration failed")
	r.ErrorContains(err, r.errMockText)
}

func (r *RegisterTestSuite) Test_500_credential_password_generation_during_registration_fails() {
	r.api.PasswordGenerator = password.NewMockGenerator("", errors.New(r.errMockText))

	acc := service.NewBaseAccount(r.user, time.Now())
	r.accounts.EXPECT().Create(context.Background(), r.user).Times(1).Return(acc, nil)
	err := r.Register(REST.RegisterParams{})

	r.ErrorContains(err, "generating a password for registration failed")
	r.ErrorContains(err, r.errMockText)
}

func (r *RegisterTestSuite) Test_500_account_registration_fails() {
	r.accounts.EXPECT().Create(r.ctx.Request().Context(), r.user).Times(1).
		Return(nil, errors.New(r.errMockText))

	err := r.Register(REST.RegisterParams{})

	r.ErrorContains(err, r.errMockText)
}

func (r *RegisterTestSuite) Test_403_account_not_exists_provided_credentials() {
	r.accounts.EXPECT().Find(r.ctx.Request().Context(), "foo").Times(1).
		Return(nil, errors.New(r.errMockText))
	r.ctx.Request().SetBasicAuth("foo", "bar")

	err := r.Register(REST.RegisterParams{})

	r.ErrorContains(err, r.errMockText)
}

func (r *RegisterTestSuite) Test_403_account_not_exists_wrong_pass() {
	acc := service.NewBaseAccount("foo", time.Now())
	r.accounts.EXPECT().Find(r.ctx.Request().Context(), "foo").Times(1).
		Return(acc, nil)
	r.devices.EXPECT().GetDevice(r.ctx.Request().Context(), acc, gomock.Any()).Times(1).
		Return(service.NewBaseDevice(r.deviceID, "test"), nil)
	r.ctx.Request().SetBasicAuth("foo", "WRONG")
	err := r.Register(REST.RegisterParams{})

	r.ErrorIs(err, v1.ErrPasswordMismatch)
}

func (r *RegisterTestSuite) Test_403_device_not_registered_no_share_code() {
	acc := service.NewBaseAccount(r.user, time.Now())

	r.ctx.Request().SetBasicAuth(r.user, r.pass)
	r.accounts.EXPECT().Find(r.ctx.Request().Context(), r.user).Times(1).
		Return(acc, nil)
	r.devices.EXPECT().GetDevice(r.ctx.Request().Context(), acc, r.deviceID).Times(1).
		Return(nil, service.ErrDeviceNotFound)

	err := r.Register(
		REST.RegisterParams{
			XDeviceID: REST.XDeviceID(r.deviceID),
		},
	)

	r.ErrorIs(err, v1.ErrDeviceNotRegistered)
}

func (r *RegisterTestSuite) Test_403_device_not_registered_share_code_lookup_fails() {
	acc := service.NewBaseAccount(r.user, time.Now())

	r.ctx.Request().SetBasicAuth(r.user, r.pass)
	r.accounts.EXPECT().Find(r.ctx.Request().Context(), r.user).Times(1).
		Return(acc, nil)
	r.devices.EXPECT().GetDevice(r.ctx.Request().Context(), acc, r.deviceID).Times(1).
		Return(nil, service.ErrDeviceNotFound)
	r.sharing.EXPECT().IsShared(r.ctx.Request().Context(), acc, r.share).Times(1).
		Return(errors.New(r.errMockText))

	err := r.Register(
		REST.RegisterParams{
			XDeviceID: REST.XDeviceID(r.deviceID),
			Share:     (*REST.ShareCode)(&r.share),
		},
	)

	r.ErrorContains(err, "cannot verify share")
}

func (r *RegisterTestSuite) Test_403_device_not_registered_share_code_invalid() {
	acc := service.NewBaseAccount(r.user, time.Now())

	r.ctx.Request().SetBasicAuth(r.user, r.pass)
	r.accounts.EXPECT().Find(r.ctx.Request().Context(), r.user).Times(1).
		Return(acc, nil)
	r.devices.EXPECT().GetDevice(r.ctx.Request().Context(), acc, r.deviceID).Times(1).
		Return(nil, service.ErrDeviceNotFound)
	r.sharing.EXPECT().IsShared(r.ctx.Request().Context(), acc, r.share).Times(1).
		Return(service.ErrShareCodeInvalid)

	err := r.Register(
		REST.RegisterParams{
			XDeviceID: REST.XDeviceID(r.deviceID),
			Share:     (*REST.ShareCode)(&r.share),
		},
	)

	r.ErrorContains(err, "is invalid (not shared)")
}

func (r *RegisterTestSuite) Test_500_device_not_registered_share_code_ok_device_registration_fails() {
	acc := service.NewBaseAccount(r.user, time.Now())

	r.ctx.Request().SetBasicAuth(r.user, r.pass)
	r.accounts.EXPECT().Find(r.ctx.Request().Context(), r.user).Times(1).
		Return(acc, nil)
	r.devices.EXPECT().GetDevice(r.ctx.Request().Context(), acc, r.deviceID).Times(1).
		Return(nil, service.ErrDeviceNotFound)
	r.sharing.EXPECT().IsShared(r.ctx.Request().Context(), acc, r.share).Times(1).
		Return(nil)
	r.devices.EXPECT().AddDevice(r.ctx.Request().Context(), acc, r.deviceID, r.pass).Times(1).
		Return(nil, errors.New(r.errMockText))

	err := r.Register(
		REST.RegisterParams{
			XDeviceID: REST.XDeviceID(r.deviceID),
			Share:     (*REST.ShareCode)(&r.share),
		},
	)

	r.ErrorContains(err, "cannot register device")
}

//nolint:lll
func (r *RegisterTestSuite) Test_500_device_not_registered_share_code_ok_device_registration_ok_share_revocation_failed() {
	acc := service.NewBaseAccount(r.user, time.Now())

	r.ctx.Request().SetBasicAuth(r.user, r.pass)
	r.accounts.EXPECT().Find(r.ctx.Request().Context(), r.user).Times(1).
		Return(acc, nil)
	r.devices.EXPECT().GetDevice(r.ctx.Request().Context(), acc, r.deviceID).Times(1).
		Return(nil, service.ErrDeviceNotFound)
	r.sharing.EXPECT().IsShared(r.ctx.Request().Context(), acc, r.share).Times(1).
		Return(nil)
	r.devices.EXPECT().AddDevice(r.ctx.Request().Context(), acc, r.deviceID, r.pass).Times(1).
		Return(service.NewBaseDevice(r.deviceID, HashedPassword(r.pass)), nil)
	r.sharing.EXPECT().Revoke(r.ctx.Request().Context(), acc, r.share).Times(1).
		Return(errors.New(r.errMockText))

	err := r.Register(
		REST.RegisterParams{
			XDeviceID: REST.XDeviceID(r.deviceID),
			Share:     (*REST.ShareCode)(&r.share),
		},
	)

	r.ErrorContains(err, r.errMockText)
}

func (r *RegisterTestSuite) Test_200_device_not_registered_share_code_ok_device_registration_ok_with_header() {
	acc := service.NewBaseAccount(r.user, time.Now())

	r.ctx.Request().SetBasicAuth(r.user, r.pass)
	r.accounts.EXPECT().Find(r.ctx.Request().Context(), r.user).Times(1).
		Return(acc, nil)
	r.devices.EXPECT().GetDevice(r.ctx.Request().Context(), acc, r.deviceID).Times(1).
		Return(nil, service.ErrDeviceNotFound)
	r.sharing.EXPECT().IsShared(r.ctx.Request().Context(), acc, r.share).Times(1).
		Return(nil)
	r.devices.EXPECT().AddDevice(r.ctx.Request().Context(), acc, r.deviceID, r.pass).Times(1).
		Return(service.NewBaseDevice(r.deviceID, HashedPassword(r.pass)), nil)
	r.sharing.EXPECT().Revoke(r.ctx.Request().Context(), acc, r.share).Times(1).
		Return(nil)

	err := r.Register(
		REST.RegisterParams{
			XDeviceID: REST.XDeviceID(r.deviceID),
			Share:     (*REST.ShareCode)(&r.share),
		},
	)

	r.NoError(err)
}

//nolint:lll
func (r *RegisterTestSuite) Test_200_device_not_registered_share_code_ok_device_registration_ok_with_generated_creds_new_account() {
	acc := service.NewBaseAccount(r.user, time.Now())

	r.accounts.EXPECT().Create(r.ctx.Request().Context(), r.user).Times(1).
		Return(acc, nil)
	r.devices.EXPECT().AddDevice(r.ctx.Request().Context(), acc, r.deviceID, r.pass).Times(1).
		Return(service.NewBaseDevice(r.deviceID, HashedPassword(r.pass)), nil)

	err := r.Register(
		REST.RegisterParams{
			XDeviceID: REST.XDeviceID(r.deviceID),
			Share:     (*REST.ShareCode)(&r.share),
		},
	)

	r.NoError(err)
}

func (r *RegisterTestSuite) Test_200_device_not_registered_share_code_ok_device_registration_ok_with_generated_creds() {
	acc := service.NewBaseAccount(r.user, time.Now())
	newPass := "some new password"

	r.accounts.EXPECT().Find(r.ctx.Request().Context(), r.user).Times(1).
		Return(acc, nil)
	r.devices.EXPECT().GetDevice(r.ctx.Request().Context(), acc, r.deviceID).Times(1).
		Return(nil, service.ErrDeviceNotFound)
	r.sharing.EXPECT().IsShared(r.ctx.Request().Context(), acc, r.share).Times(1).
		Return(nil)
	r.devices.EXPECT().AddDevice(r.ctx.Request().Context(), acc, r.deviceID, newPass).Times(1).
		Return(service.NewBaseDevice(r.deviceID, HashedPassword(newPass)), nil)
	r.sharing.EXPECT().Revoke(r.ctx.Request().Context(), acc, r.share).Times(1).
		Return(nil)

	r.ctx.Request().SetBasicAuth(r.user, newPass)
	err := r.Register(
		REST.RegisterParams{
			XDeviceID: REST.XDeviceID(r.deviceID),
			Share:     (*REST.ShareCode)(&r.share),
		},
	)

	r.NoError(err)
}
