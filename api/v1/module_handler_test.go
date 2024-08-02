package v1_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	json "github.com/json-iterator/go"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	v1 "github.com/jakobmoellerdev/octi-sync-server/api/v1"
	"github.com/jakobmoellerdev/octi-sync-server/api/v1/REST"
	"github.com/jakobmoellerdev/octi-sync-server/middleware/basic"
	"github.com/jakobmoellerdev/octi-sync-server/middleware/logging"
	"github.com/jakobmoellerdev/octi-sync-server/service"
	"github.com/jakobmoellerdev/octi-sync-server/service/mock"
	"github.com/jakobmoellerdev/octi-sync-server/service/redis"
)

const (
	moduleName = "test"
	moduleData = "test"
	username   = "username"
)

type ModuleTestSuite struct {
	suite.Suite
	api      *v1.API
	modules  *mock.MockModules
	metadata *mock.MockMetadataProvider
	devices  *mock.MockDevices
	server   *echo.Echo
	deviceID uuid.UUID
	rec      *httptest.ResponseRecorder
	user     *mock.MockAccount
	device   *mock.MockDevice
}

func TestModuleTestSuite(t *testing.T) {
	suite.Run(t, &ModuleTestSuite{})
}

func (m *ModuleTestSuite) SetupTest() {
	ctrl := gomock.NewController(m.T())
	m.modules = mock.NewMockModules(ctrl)
	m.metadata = mock.NewMockMetadataProvider(ctrl)
	m.devices = mock.NewMockDevices(ctrl)
	m.server = echo.New()
	logger := zerolog.New(zerolog.NewConsoleWriter(zerolog.ConsoleTestWriter(m.T())))
	m.server.Use(logging.RequestLogging(&logger))
	m.api = &v1.API{
		Modules:          m.modules,
		MetadataProvider: m.metadata,
		Devices:          m.devices,
	}
	deviceID, err := uuid.NewRandom()

	m.NoError(err)

	m.deviceID = deviceID
	m.rec = httptest.NewRecorder()
	m.user = mock.NewMockAccount(ctrl)
	m.device = mock.NewMockDevice(ctrl)

	m.device.EXPECT().ID().AnyTimes().Return(service.DeviceID(m.deviceID))
	m.user.EXPECT().Username().AnyTimes().Return(username)
}

func (m *ModuleTestSuite) TestAPI_CreateModule() {
	req := emptyRequest(http.MethodPost)

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	ctx := m.server.NewContext(req, m.rec)

	ctx.Set(basic.AccountKey, m.user)
	ctx.Set(basic.Device, m.device)

	m.modules.EXPECT().Set(
		ctx.Request().Context(), fmt.Sprintf("%s-%s-%s", m.user.Username(), m.deviceID, moduleName),
		gomock.Any(),
	).Return(nil)

	m.metadata.EXPECT().Set(ctx.Request().Context(), gomock.Any()).Return(nil)

	if m.NoError(
		m.api.CreateModule(ctx, moduleName, REST.CreateModuleParams{XDeviceID: m.deviceID}),
	) {
		m.Equal(http.StatusAccepted, m.rec.Code)
		m.NoError(json.Unmarshal(m.rec.Body.Bytes(), &map[string]string{}))
	}
}

func (m *ModuleTestSuite) TestAPI_CreateModule_WriteFails() {
	req := emptyRequest(http.MethodPost)

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	ctx := m.server.NewContext(req, m.rec)

	ctx.Set(basic.AccountKey, m.user)
	ctx.Set(basic.Device, service.NewBaseDevice(service.DeviceID(m.deviceID), "test"))

	m.modules.EXPECT().Set(
		ctx.Request().Context(), fmt.Sprintf("%s-%s-%s", m.user.Username(), m.deviceID, moduleName),
		gomock.Any(),
	).Return(errors.New("set error"))

	m.ErrorContains(
		m.api.CreateModule(ctx, moduleName, REST.CreateModuleParams{XDeviceID: m.deviceID}),
		"could not create/update module",
	)
}

func (m *ModuleTestSuite) TestAPI_CreateModule_MetadataFails() {
	req := emptyRequest(http.MethodPost)

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	ctx := m.server.NewContext(req, m.rec)

	ctx.Set(basic.AccountKey, m.user)
	ctx.Set(basic.Device, m.device)

	m.modules.EXPECT().Set(
		ctx.Request().Context(), fmt.Sprintf("%s-%s-%s", m.user.Username(), m.deviceID, moduleName),
		gomock.Any(),
	).Return(nil)

	m.metadata.EXPECT().Set(ctx.Request().Context(), gomock.Any()).Return(
		errors.New("metadata set err"),
	)

	m.ErrorContains(
		m.api.CreateModule(ctx, moduleName, REST.CreateModuleParams{XDeviceID: m.deviceID}),
		"could not create/update module metadata",
	)
}

func (m *ModuleTestSuite) TestAPI_CreateModule_NoAccount() {
	req := emptyRequest(http.MethodPost)

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	ctx := m.server.NewContext(req, m.rec)

	m.Error(
		m.api.GetModule(ctx, moduleName, REST.GetModuleParams{XDeviceID: m.deviceID}),
		v1.ErrAccountForVerifyingDeviceNotPresent,
	)
}

func (m *ModuleTestSuite) TestAPI_GetModule_NoAccount() {
	req := emptyRequest(http.MethodGet)

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	ctx := m.server.NewContext(req, m.rec)

	m.Error(
		m.api.GetModule(ctx, moduleName, REST.GetModuleParams{XDeviceID: m.deviceID}),
		v1.ErrAccountForVerifyingDeviceNotPresent,
	)
}

func (m *ModuleTestSuite) TestAPI_GetModule() {
	req := emptyRequest(http.MethodPost)

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	ctx := m.server.NewContext(req, m.rec)

	ctx.Set(basic.AccountKey, m.user)
	ctx.Set(basic.Device, m.device)

	id := fmt.Sprintf("%s-%s-%s", m.user.Username(), m.deviceID, moduleName)

	m.modules.EXPECT().Get(
		ctx.Request().Context(), id,
	).Return(redis.ModuleFromBytes([]byte(moduleData)), nil)

	m.metadata.EXPECT().Get(ctx.Request().Context(), service.MetadataID(id)).Return(
		service.NewBaseMetadata(
			id, time.Now(),
		), nil,
	)

	if m.NoError(
		m.api.GetModule(ctx, moduleName, REST.GetModuleParams{XDeviceID: m.deviceID}),
	) {
		m.Equal(http.StatusOK, m.rec.Code)

		body := m.rec.Body.String()

		m.NotEmpty(body)
		m.Equal(moduleData, body)
	}
}

func (m *ModuleTestSuite) TestAPI_GetModule_By_Param() {
	for _, testCase := range []struct {
		dataInReturn []byte
		expectedCode int
	}{
		{[]byte(moduleData), http.StatusOK},
		{[]byte{}, http.StatusNoContent},
	} {
		secondDeviceId := uuid.Must(uuid.NewRandom())
		secondDevice := mock.NewMockDevice(gomock.NewController(m.T()))
		secondDevice.EXPECT().ID().Return(service.DeviceID(secondDeviceId))

		m.rec = httptest.NewRecorder()
		req := emptyRequest(http.MethodPost)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		ctx := m.server.NewContext(req, m.rec)
		id := fmt.Sprintf("%s-%s-%s", m.user.Username(), secondDeviceId, moduleName)

		ctx.Set(basic.AccountKey, m.user)
		ctx.Set(basic.Device, m.device)

		m.devices.EXPECT().GetDevice(ctx.Request().Context(), m.user, service.DeviceID(secondDeviceId)).
			Return(secondDevice, nil)

		if len(testCase.dataInReturn) > 0 {
			m.metadata.EXPECT().Get(ctx.Request().Context(), service.MetadataID(id)).Return(
				service.NewBaseMetadata(
					id, time.Now(),
				), nil,
			)
		}

		m.modules.EXPECT().Get(
			ctx.Request().Context(), id,
		).Return(redis.ModuleFromBytes(testCase.dataInReturn), nil)

		if m.NoError(
			m.api.GetModule(
				ctx, moduleName, REST.GetModuleParams{
					XDeviceID: m.deviceID,
					DeviceId:  &secondDeviceId,
				},
			),
		) {
			m.Equal(testCase.expectedCode, m.rec.Code)
			body := m.rec.Body.String()
			m.Equal(string(testCase.dataInReturn), body)
		}
	}
}
