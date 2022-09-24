package basic_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	auth "github.com/jakob-moeller-cloud/octi-sync-server/middleware/basic"
	"github.com/jakob-moeller-cloud/octi-sync-server/middleware/logging"
	"github.com/jakob-moeller-cloud/octi-sync-server/service"
	"github.com/jakob-moeller-cloud/octi-sync-server/service/memory"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/random"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context.
type BasicAuthTestSuite struct {
	suite.Suite
	devices  service.Devices
	accounts service.Accounts
	echo.Context
	req *http.Request
	api *echo.Echo
}

func (suite *BasicAuthTestSuite) SetupSuite() {
	logger := zerolog.New(zerolog.NewTestWriter(suite.T()))
	api := echo.New()
	api.Use(logging.RequestLogging(&logger))
	suite.api = api

	suite.accounts, suite.devices = memory.NewAccounts(), memory.NewDevices()
}

func (suite *BasicAuthTestSuite) SetupTest() {
	suite.req = httptest.NewRequest(http.MethodGet, "/some-resource", nil)
	suite.Context = suite.api.NewContext(suite.req, httptest.NewRecorder())
}

func (suite *BasicAuthTestSuite) ResetRequest() {
	suite.Context.Reset(suite.req, suite.Context.Response().Writer)
}

func (suite *BasicAuthTestSuite) registerAndSetAuthorizationHeader(user string) service.Account {
	acc, pass, err := suite.accounts.Register(context.Background(), user)

	suite.NoError(err)
	suite.req.SetBasicAuth(user, pass)

	return acc
}

func (suite *BasicAuthTestSuite) registerAndSetDeviceHeader(acc service.Account, deviceID string) service.Device {
	dev, err := suite.devices.Register(context.Background(), acc, deviceID)

	suite.NoError(err)
	suite.req.Header.Set(auth.DeviceIDHeader, deviceID)

	return dev
}

func (suite *BasicAuthTestSuite) randomUsername() string { return "test-user-" + random.String(5) }
func (suite *BasicAuthTestSuite) randomDeviceID() string { return "device-" + random.String(5) }

//goland:noinspection SpellCheckingInspection
func (suite *BasicAuthTestSuite) asHTTPError(err error) *echo.HTTPError {
	httpError, ok := err.(*echo.HTTPError) //nolint:errorlint

	suite.True(ok, "When using Middleware Error Assertion the err has to be *echo.HTTPError")

	return httpError
}

// All methods that begin with "Test" are run as tests within a
// suite.
func (suite *BasicAuthTestSuite) TestAuthWithSharing() {
	testMiddleware := auth.AuthWithShare(suite.accounts, suite.devices)(http200)

	var (
		acc service.Account
		dev service.Device
	)

	// No Token
	suite.Equal(http.StatusUnauthorized, suite.asHTTPError(testMiddleware(suite)).Code)

	// Valid Token
	acc = suite.registerAndSetAuthorizationHeader(suite.randomUsername())
	dev = suite.registerAndSetDeviceHeader(acc, suite.randomDeviceID())
	suite.NoError(testMiddleware(suite))
	suite.ResetRequest()

	// Trying to access account from random other device
	suite.req.Header.Set(auth.DeviceIDHeader, suite.randomDeviceID())
	suite.Equal(http.StatusForbidden, suite.asHTTPError(testMiddleware(suite)).Code)
	suite.ResetRequest()

	// switching back to normal device should be okay
	suite.req.Header.Set(auth.DeviceIDHeader, dev.ID())
	suite.NoError(testMiddleware(suite))
	suite.ResetRequest()

	// using a random authorization header with a registered device is blocked immediately, as
	// there is no account
	suite.req.Header.Set(echo.HeaderAuthorization, "")
	suite.Equal(http.StatusUnauthorized, suite.asHTTPError(testMiddleware(suite)).Code)
	suite.ResetRequest()

	// Valid Token with both params reset should work fine
	acc = suite.registerAndSetAuthorizationHeader(suite.randomUsername())
	suite.registerAndSetDeviceHeader(acc, suite.randomDeviceID())
	suite.NoError(testMiddleware(suite))
	suite.ResetRequest()

	// Now we share an account
	share, err := suite.accounts.Share(context.Background(), acc.Username())
	suite.NoError(err)
	// at first the device is not shared, the call should be forbidden
	suite.req.Header.Set(auth.DeviceIDHeader, suite.randomDeviceID())
	suite.Equal(http.StatusForbidden, suite.asHTTPError(testMiddleware(suite)).Code)
	suite.ResetRequest()

	// now it should still be invalid even though we provide a share code (that was not okay)
	setQuery(suite.req.URL, auth.ShareQueryParamName, url.QueryEscape("invalid"))
	suite.Equal(http.StatusForbidden, suite.asHTTPError(testMiddleware(suite)).Code)
	suite.ResetRequest()

	// now it should be valid as we provide a valid share code
	setQuery(suite.req.URL, auth.ShareQueryParamName, url.QueryEscape(share))
	suite.NoError(testMiddleware(suite))
	suite.ResetRequest()
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run.
func TestExampleTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(BasicAuthTestSuite))
}

func setQuery(u *url.URL, key string, values ...string) {
	u.RawQuery = url.Values{key: values}.Encode()
}

func http200(context echo.Context) error {
	return context.String(http.StatusOK, "test") //nolint:wrapcheck
}
