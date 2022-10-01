package v1_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	v1 "github.com/jakob-moeller-cloud/octi-sync-server/api/v1"
	"github.com/jakob-moeller-cloud/octi-sync-server/api/v1/REST"
	"github.com/jakob-moeller-cloud/octi-sync-server/middleware/basic"
	"github.com/jakob-moeller-cloud/octi-sync-server/middleware/logging"
	"github.com/jakob-moeller-cloud/octi-sync-server/service"
	"github.com/jakob-moeller-cloud/octi-sync-server/service/mock"
	"github.com/jakob-moeller-cloud/octi-sync-server/service/redis"
	json "github.com/json-iterator/go"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestAPI_CreateModule(t *testing.T) {
	t.Parallel()
	logger := zerolog.New(zerolog.NewConsoleWriter(zerolog.ConsoleTestWriter(t)))
	server := echo.New()
	server.Use(logging.RequestLogging(&logger))
	assertions := assert.New(t)
	ctrl := gomock.NewController(t)
	modules := mock.NewMockModules(ctrl)
	metadata := mock.NewMockMetadataProvider(ctrl)
	api := &v1.API{
		Modules:          modules,
		MetadataProvider: metadata,
	}

	deviceID, err := uuid.NewRandom()
	assertions.NoError(err)

	rec := httptest.NewRecorder()
	req := emptyRequest(http.MethodPost)

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	ctx := server.NewContext(req, rec)

	user := mock.NewMockAccount(ctrl)
	user.EXPECT().Username().AnyTimes().Return("username")
	ctx.Set(basic.AccountKey, user)

	moduleName := "test"

	modules.EXPECT().Set(
		ctx.Request().Context(), fmt.Sprintf("%s-%s-%s", user.Username(), deviceID, moduleName),
		gomock.Any(),
	).Return(nil)

	metadata.EXPECT().Set(ctx.Request().Context(), gomock.Any()).Return(nil)

	if assertions.NoError(
		api.CreateModule(ctx, moduleName, REST.CreateModuleParams{XDeviceID: deviceID}),
	) {
		verifyCreateModuleResponse(assertions, rec)
	}
}

func verifyCreateModuleResponse(assert *assert.Assertions, rec *httptest.ResponseRecorder) {
	assert.Equal(http.StatusAccepted, rec.Code)
	assert.NoError(json.Unmarshal(rec.Body.Bytes(), &map[string]string{}))
}

func TestAPI_GetModule(t *testing.T) {
	t.Parallel()
	logger := zerolog.New(zerolog.NewConsoleWriter(zerolog.ConsoleTestWriter(t)))
	server := echo.New()
	server.Use(logging.RequestLogging(&logger))
	assertions := assert.New(t)
	ctrl := gomock.NewController(t)
	modules := mock.NewMockModules(ctrl)
	metadata := mock.NewMockMetadataProvider(ctrl)
	apiImpl := &v1.API{
		Modules:          modules,
		MetadataProvider: metadata,
	}

	req := emptyRequest(http.MethodPost)

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	moduleName := "test"
	deviceID, err := uuid.NewRandom()
	assertions.NoError(err)

	rec := httptest.NewRecorder()
	ctx := server.NewContext(req, rec)

	user := mock.NewMockAccount(ctrl)
	user.EXPECT().Username().AnyTimes().Return("username")
	ctx.Set(basic.AccountKey, user)
	id := fmt.Sprintf("%s-%s-%s", user.Username(), deviceID, moduleName)
	modules.EXPECT().Get(
		ctx.Request().Context(), id,
	).Return(redis.ModuleFromBytes([]byte("test")), nil)

	metadata.EXPECT().Get(ctx.Request().Context(), service.MetadataID(id)).Return(
		service.NewBaseMetadata(
			id, time.Now(),
		), nil,
	)

	if assertions.NoError(
		apiImpl.GetModule(ctx, moduleName, REST.GetModuleParams{XDeviceID: deviceID}),
	) {
		verifyGetModuleResponse(assertions, rec)
	}
}

func verifyGetModuleResponse(assert *assert.Assertions, rec *httptest.ResponseRecorder) {
	assert.Equal(http.StatusOK, rec.Code)

	body := rec.Body.String()
	assert.NotEmpty(body)

	assert.Equal("test", body)
}
