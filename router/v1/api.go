package v1

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jakob-moeller-cloud/octi-sync-server/config"
	"github.com/jakob-moeller-cloud/octi-sync-server/service"

	authmiddleware "github.com/jakob-moeller-cloud/octi-sync-server/middleware/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

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
	}
}

type ModuleRequest struct {
	Name string `uri:"name" binding:"required"`
}

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

func getModule(modules service.Modules) gin.HandlerFunc {
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
