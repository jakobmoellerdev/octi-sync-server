// Package REST provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.3.0 DO NOT EDIT.
package REST

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
	"github.com/oapi-codegen/runtime"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

const (
	DeviceAuthScopes = "deviceAuth.Scopes"
)

// Defines values for HealthResult.
const (
	Down HealthResult = "Down"
	Up   HealthResult = "Up"
)

// Device a device
type Device struct {
	// Id Device ID is the unique identifier for a remote device
	Id DeviceID `json:"id"`
}

// DeviceID Device ID is the unique identifier for a remote device
type DeviceID = openapi_types.UUID

// DeviceList list of devices
type DeviceList struct {
	// Count Amount of Items contained in List
	Count ListItemCount `json:"count"`

	// Items array of devices, it will always at least contain the device of the authenticated user
	Items []Device `json:"items"`
}

// HealthAggregation defines model for HealthAggregation.
type HealthAggregation struct {
	// Components The different Components of the Server
	Components *[]HealthAggregationComponent `json:"components,omitempty"`

	// Health A Health Check Result
	Health HealthResult `json:"health"`
}

// HealthAggregationComponent defines model for HealthAggregationComponent.
type HealthAggregationComponent struct {
	// Health A Health Check Result
	Health HealthResult `json:"health"`

	// Name The Name of the Component to be Health Checked
	Name string `json:"name"`
}

// HealthResult A Health Check Result
type HealthResult string

// ListItemCount Amount of Items contained in List
type ListItemCount = int

// ModifiedAtTimestamp A Timestamp indicating when a datum was last modified
type ModifiedAtTimestamp = time.Time

// ModuleDataStream Module Data Stream
type ModuleDataStream = openapi_types.File

// ModuleName Module Name
type ModuleName = string

// RegistrationResult defines model for RegistrationResult.
type RegistrationResult struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

// ShareResponse defines model for ShareResponse.
type ShareResponse struct {
	ShareCode *string `json:"shareCode,omitempty"`
}

// DeviceIDQuery Device ID is the unique identifier for a remote device
type DeviceIDQuery = DeviceID

// ShareCode defines model for ShareCode.
type ShareCode = string

// XDeviceID Device ID is the unique identifier for a remote device
type XDeviceID = DeviceID

// DeviceListResponse list of devices
type DeviceListResponse = DeviceList

// ModuleDataAccepted An Empty JSON
type ModuleDataAccepted = interface{}

// ModuleDeletionAccepted An Empty JSON
type ModuleDeletionAccepted = interface{}

// RegisterParams defines parameters for Register.
type RegisterParams struct {
	// Share The Share Code from the Share API. If presented in combination with a new Device ID,
	// it can be used to add new devices to an account.
	Share *ShareCode `form:"share,omitempty" json:"share,omitempty"`

	// XDeviceID Unique Identifier of the calling Device. If calling Data endpoints, must be presented in order
	// to be properly authenticated.
	XDeviceID XDeviceID `json:"X-Device-ID"`
}

// ShareParams defines parameters for Share.
type ShareParams struct {
	// XDeviceID Unique Identifier of the calling Device. If calling Data endpoints, must be presented in order
	// to be properly authenticated.
	XDeviceID XDeviceID `json:"X-Device-ID"`
}

// GetDevicesParams defines parameters for GetDevices.
type GetDevicesParams struct {
	// XDeviceID Unique Identifier of the calling Device. If calling Data endpoints, must be presented in order
	// to be properly authenticated.
	XDeviceID XDeviceID `json:"X-Device-ID"`
}

// DeleteModulesParams defines parameters for DeleteModules.
type DeleteModulesParams struct {
	// DeviceId Device Identifier to use for the Query. If given, takes precedence over X-Device-ID or other hints.
	// Use to query data from devices in your account from another account.
	DeviceId *DeviceIDQuery `form:"device-id,omitempty" json:"device-id,omitempty"`

	// XDeviceID Unique Identifier of the calling Device. If calling Data endpoints, must be presented in order
	// to be properly authenticated.
	XDeviceID XDeviceID `json:"X-Device-ID"`
}

// GetModuleParams defines parameters for GetModule.
type GetModuleParams struct {
	// DeviceId Device Identifier to use for the Query. If given, takes precedence over X-Device-ID or other hints.
	// Use to query data from devices in your account from another account.
	DeviceId *DeviceIDQuery `form:"device-id,omitempty" json:"device-id,omitempty"`

	// XDeviceID Unique Identifier of the calling Device. If calling Data endpoints, must be presented in order
	// to be properly authenticated.
	XDeviceID XDeviceID `json:"X-Device-ID"`
}

// CreateModuleParams defines parameters for CreateModule.
type CreateModuleParams struct {
	// XDeviceID Unique Identifier of the calling Device. If calling Data endpoints, must be presented in order
	// to be properly authenticated.
	XDeviceID XDeviceID `json:"X-Device-ID"`
}

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Register A Device
	// (POST /auth/register)
	Register(ctx echo.Context, params RegisterParams) error
	// Share your Account
	// (POST /auth/share)
	Share(ctx echo.Context, params ShareParams) error
	// Get All registered Devices for your Account
	// (GET /devices)
	GetDevices(ctx echo.Context, params GetDevicesParams) error
	// Checks if the Service is Available for Processing Request
	// (GET /health)
	IsHealthy(ctx echo.Context) error
	// Clears Module Data for a Device
	// (DELETE /module)
	DeleteModules(ctx echo.Context, params DeleteModulesParams) error
	// Get Module Data
	// (GET /module/{name})
	GetModule(ctx echo.Context, name ModuleName, params GetModuleParams) error
	// Create/Update Module Data
	// (POST /module/{name})
	CreateModule(ctx echo.Context, name ModuleName, params CreateModuleParams) error
	// Checks if the Service is Operational
	// (GET /ready)
	IsReady(ctx echo.Context) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// Register converts echo context to params.
func (w *ServerInterfaceWrapper) Register(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params RegisterParams
	// ------------- Optional query parameter "share" -------------

	err = runtime.BindQueryParameter("form", true, false, "share", ctx.QueryParams(), &params.Share)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter share: %s", err))
	}

	headers := ctx.Request().Header
	// ------------- Required header parameter "X-Device-ID" -------------
	if valueList, found := headers[http.CanonicalHeaderKey("X-Device-ID")]; found {
		var XDeviceID XDeviceID
		n := len(valueList)
		if n != 1 {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Expected one value for X-Device-ID, got %d", n))
		}

		err = runtime.BindStyledParameterWithOptions("simple", "X-Device-ID", valueList[0], &XDeviceID, runtime.BindStyledParameterOptions{ParamLocation: runtime.ParamLocationHeader, Explode: false, Required: true})
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter X-Device-ID: %s", err))
		}

		params.XDeviceID = XDeviceID
	} else {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Header parameter X-Device-ID is required, but not found"))
	}

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.Register(ctx, params)
	return err
}

// Share converts echo context to params.
func (w *ServerInterfaceWrapper) Share(ctx echo.Context) error {
	var err error

	ctx.Set(DeviceAuthScopes, []string{})

	// Parameter object where we will unmarshal all parameters from the context
	var params ShareParams

	headers := ctx.Request().Header
	// ------------- Required header parameter "X-Device-ID" -------------
	if valueList, found := headers[http.CanonicalHeaderKey("X-Device-ID")]; found {
		var XDeviceID XDeviceID
		n := len(valueList)
		if n != 1 {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Expected one value for X-Device-ID, got %d", n))
		}

		err = runtime.BindStyledParameterWithOptions("simple", "X-Device-ID", valueList[0], &XDeviceID, runtime.BindStyledParameterOptions{ParamLocation: runtime.ParamLocationHeader, Explode: false, Required: true})
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter X-Device-ID: %s", err))
		}

		params.XDeviceID = XDeviceID
	} else {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Header parameter X-Device-ID is required, but not found"))
	}

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.Share(ctx, params)
	return err
}

// GetDevices converts echo context to params.
func (w *ServerInterfaceWrapper) GetDevices(ctx echo.Context) error {
	var err error

	ctx.Set(DeviceAuthScopes, []string{})

	// Parameter object where we will unmarshal all parameters from the context
	var params GetDevicesParams

	headers := ctx.Request().Header
	// ------------- Required header parameter "X-Device-ID" -------------
	if valueList, found := headers[http.CanonicalHeaderKey("X-Device-ID")]; found {
		var XDeviceID XDeviceID
		n := len(valueList)
		if n != 1 {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Expected one value for X-Device-ID, got %d", n))
		}

		err = runtime.BindStyledParameterWithOptions("simple", "X-Device-ID", valueList[0], &XDeviceID, runtime.BindStyledParameterOptions{ParamLocation: runtime.ParamLocationHeader, Explode: false, Required: true})
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter X-Device-ID: %s", err))
		}

		params.XDeviceID = XDeviceID
	} else {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Header parameter X-Device-ID is required, but not found"))
	}

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.GetDevices(ctx, params)
	return err
}

// IsHealthy converts echo context to params.
func (w *ServerInterfaceWrapper) IsHealthy(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.IsHealthy(ctx)
	return err
}

// DeleteModules converts echo context to params.
func (w *ServerInterfaceWrapper) DeleteModules(ctx echo.Context) error {
	var err error

	ctx.Set(DeviceAuthScopes, []string{})

	// Parameter object where we will unmarshal all parameters from the context
	var params DeleteModulesParams
	// ------------- Optional query parameter "device-id" -------------

	err = runtime.BindQueryParameter("form", true, false, "device-id", ctx.QueryParams(), &params.DeviceId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter device-id: %s", err))
	}

	headers := ctx.Request().Header
	// ------------- Required header parameter "X-Device-ID" -------------
	if valueList, found := headers[http.CanonicalHeaderKey("X-Device-ID")]; found {
		var XDeviceID XDeviceID
		n := len(valueList)
		if n != 1 {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Expected one value for X-Device-ID, got %d", n))
		}

		err = runtime.BindStyledParameterWithOptions("simple", "X-Device-ID", valueList[0], &XDeviceID, runtime.BindStyledParameterOptions{ParamLocation: runtime.ParamLocationHeader, Explode: false, Required: true})
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter X-Device-ID: %s", err))
		}

		params.XDeviceID = XDeviceID
	} else {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Header parameter X-Device-ID is required, but not found"))
	}

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.DeleteModules(ctx, params)
	return err
}

// GetModule converts echo context to params.
func (w *ServerInterfaceWrapper) GetModule(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "name" -------------
	var name ModuleName

	err = runtime.BindStyledParameterWithOptions("simple", "name", ctx.Param("name"), &name, runtime.BindStyledParameterOptions{ParamLocation: runtime.ParamLocationPath, Explode: false, Required: true})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter name: %s", err))
	}

	ctx.Set(DeviceAuthScopes, []string{})

	// Parameter object where we will unmarshal all parameters from the context
	var params GetModuleParams
	// ------------- Optional query parameter "device-id" -------------

	err = runtime.BindQueryParameter("form", true, false, "device-id", ctx.QueryParams(), &params.DeviceId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter device-id: %s", err))
	}

	headers := ctx.Request().Header
	// ------------- Required header parameter "X-Device-ID" -------------
	if valueList, found := headers[http.CanonicalHeaderKey("X-Device-ID")]; found {
		var XDeviceID XDeviceID
		n := len(valueList)
		if n != 1 {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Expected one value for X-Device-ID, got %d", n))
		}

		err = runtime.BindStyledParameterWithOptions("simple", "X-Device-ID", valueList[0], &XDeviceID, runtime.BindStyledParameterOptions{ParamLocation: runtime.ParamLocationHeader, Explode: false, Required: true})
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter X-Device-ID: %s", err))
		}

		params.XDeviceID = XDeviceID
	} else {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Header parameter X-Device-ID is required, but not found"))
	}

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.GetModule(ctx, name, params)
	return err
}

// CreateModule converts echo context to params.
func (w *ServerInterfaceWrapper) CreateModule(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "name" -------------
	var name ModuleName

	err = runtime.BindStyledParameterWithOptions("simple", "name", ctx.Param("name"), &name, runtime.BindStyledParameterOptions{ParamLocation: runtime.ParamLocationPath, Explode: false, Required: true})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter name: %s", err))
	}

	ctx.Set(DeviceAuthScopes, []string{})

	// Parameter object where we will unmarshal all parameters from the context
	var params CreateModuleParams

	headers := ctx.Request().Header
	// ------------- Required header parameter "X-Device-ID" -------------
	if valueList, found := headers[http.CanonicalHeaderKey("X-Device-ID")]; found {
		var XDeviceID XDeviceID
		n := len(valueList)
		if n != 1 {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Expected one value for X-Device-ID, got %d", n))
		}

		err = runtime.BindStyledParameterWithOptions("simple", "X-Device-ID", valueList[0], &XDeviceID, runtime.BindStyledParameterOptions{ParamLocation: runtime.ParamLocationHeader, Explode: false, Required: true})
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter X-Device-ID: %s", err))
		}

		params.XDeviceID = XDeviceID
	} else {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Header parameter X-Device-ID is required, but not found"))
	}

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.CreateModule(ctx, name, params)
	return err
}

// IsReady converts echo context to params.
func (w *ServerInterfaceWrapper) IsReady(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.IsReady(ctx)
	return err
}

// This is a simple interface which specifies echo.Route addition functions which
// are present on both echo.Echo and echo.Group, since we want to allow using
// either of them for path registration
type EchoRouter interface {
	CONNECT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	TRACE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}

// RegisterHandlers adds each server route to the EchoRouter.
func RegisterHandlers(router EchoRouter, si ServerInterface) {
	RegisterHandlersWithBaseURL(router, si, "")
}

// Registers handlers, and prepends BaseURL to the paths, so that the paths
// can be served under a prefix.
func RegisterHandlersWithBaseURL(router EchoRouter, si ServerInterface, baseURL string) {

	wrapper := ServerInterfaceWrapper{
		Handler: si,
	}

	router.POST(baseURL+"/auth/register", wrapper.Register)
	router.POST(baseURL+"/auth/share", wrapper.Share)
	router.GET(baseURL+"/devices", wrapper.GetDevices)
	router.GET(baseURL+"/health", wrapper.IsHealthy)
	router.DELETE(baseURL+"/module", wrapper.DeleteModules)
	router.GET(baseURL+"/module/:name", wrapper.GetModule)
	router.POST(baseURL+"/module/:name", wrapper.CreateModule)
	router.GET(baseURL+"/ready", wrapper.IsReady)

}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"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",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %w", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	res := make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	resolvePath := PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		pathToFile := url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
