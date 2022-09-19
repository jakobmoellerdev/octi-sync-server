package v1

import (
	"context"
	"errors"
	"fmt"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"net/http"
	"octi-sync-server/config"
	"octi-sync-server/service"

	authmiddleware "octi-sync-server/middleware/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var ErrDeviceIDNotPropagated = errors.New("device id was not propagated from middleware")

// New creates the V1 API Group
func New(_ context.Context, engine *gin.Engine, config *config.Config) {
	v1 := engine.Group("/v1") //nolint:varnamelen

	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", register(
				config.Services.Accounts,
				config.Services.Devices,
			))
		}

		module := v1.Group("/module")
		module.Use(
			authmiddleware.BasicAuth(config.Services.Accounts),
			authmiddleware.DeviceAuth(config.Services.Devices),
		)
		{
			module.GET("/:name", getModule(config.Services.Modules))
			module.POST("/:name", createModule(config.Services.Modules))
		}
		v1.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	}
}

type ModuleRequest struct {
	Name string `uri:"name" binding:"required"`
}

// createModule godoc
// @Summary     createModule Create the Module
// @Description Create the Module
// @Accept      application/octet-stream
// @Tags        modules
// @Success     200
// @Failure     400
// @Failure     401
// @Failure     403
// @Failure     404
// @Failure     413
// @Failure     429
// @Failure     500
// @Param       name path string true "module-name"
// @Param       moduleData body string true "module-name"
// @Param       X-Device-ID  header string true "Device Header"
// @Security    BasicAuth
// @Router      /module/{name} [post]
func createModule(modules service.Modules) gin.HandlerFunc {
	return func(context *gin.Context) {
		var request ModuleRequest
		if err := context.ShouldBindUri(&request); err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"msg": err})
			return
		}

		deviceID, found := context.Get(authmiddleware.DeviceID)
		if !found {
			context.JSON(http.StatusInternalServerError, gin.H{
				"msg": "device id was not propagated from middleware",
			})
			return
		}

		err := modules.Set(
			context.Request.Context(),
			fmt.Sprintf("%s-%s", deviceID, request.Name),
			service.RedisModuleFromReader(context.Request.Body, int(context.Request.ContentLength)),
		)
		if err != nil {
			_ = context.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		context.JSON(http.StatusOK, gin.H{})
	}
}

// getModule godoc
// @Summary     getModule Get the Module
// @Description Get the Module
// @Produce     application/octet-stream
// @Tags        modules
// @Success     200 {string} binary
// @Failure     400
// @Failure     401
// @Failure     403
// @Failure     404
// @Failure     413
// @Failure     429
// @Failure     500
// @Param       name path string true "module-name"
// @Param       X-Device-ID  header string true "Device Header"
// @Security    BasicAuth
// @Router      /module/{name} [get]
func getModule(modules service.Modules) gin.HandlerFunc {
	return func(context *gin.Context) {
		var request ModuleRequest
		if err := context.ShouldBindUri(&request); err != nil {
			_ = context.AbortWithError(http.StatusBadRequest, err)
			return
		}

		deviceID, found := context.Get(authmiddleware.DeviceID)
		if !found {
			_ = context.AbortWithError(http.StatusInternalServerError, ErrDeviceIDNotPropagated)
			return
		}

		module, err := modules.Get(context.Request.Context(),
			fmt.Sprintf("%s-%s", deviceID, request.Name))
		if err != nil {
			_ = context.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		// context.Stream(func(w io.Writer) bool {
		//	if _, err := io.Copy(w, module.Raw()); err != nil {
		//		_ = context.AbortWithError(http.StatusInternalServerError, err)
		//	}
		//	return false
		// })

		context.DataFromReader(
			http.StatusOK,
			int64(module.Size()),
			"application/octet-stream",
			module.Raw(),
			map[string]string{},
		)
	}
}

type RegistrationResponse struct {
	Username string
	DeviceID string
	Password string
}

// register godoc
// @Summary     register Device Registration
// @Description Creates a new Registration
// @Tags        auth
// @Success     200
// @Failure     400
// @Failure     404
// @Failure     413
// @Failure     429
// @Failure     500
// @Param       X-Device-ID  header string false "Device Header"
// @Router      /auth/register [post]
func register(accounts service.Accounts, devices service.Devices) gin.HandlerFunc {
	return func(context *gin.Context) {
		accountID, err := uuid.NewRandom()
		if err != nil {
			_ = context.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		deviceID := context.GetHeader(authmiddleware.DeviceIDHeader)
		if deviceID == "" {
			deviceUUID, err := uuid.NewRandom()
			if err != nil {
				_ = context.AbortWithError(http.StatusInternalServerError, err)
				return
			}
			deviceID = deviceUUID.String()
			context.Header(authmiddleware.DeviceIDHeader, deviceID)
		}

		acc, password, err := accounts.Register(context.Request.Context(), accountID.String())
		if err != nil {
			_ = context.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		if err := devices.Register(context.Request.Context(), acc, deviceID); err != nil {
			_ = context.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		context.JSON(http.StatusOK, &RegistrationResponse{
			Username: acc.Username(),
			DeviceID: deviceID,
			Password: password,
		})
	}
}
