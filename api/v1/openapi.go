package v1

import (
	_ "embed" // imported for openapi specification embedding
	"fmt"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

//go:embed openapi.yaml
var openAPI []byte

type OpenAPIHandler struct {
	OpenAPI *openapi3.T
	*zerolog.Logger
	swaggerJSON []byte
}

func (h *OpenAPIHandler) ServeOpenAPI(ctx echo.Context) error {
	var err error

	switch ctx.Request().Header.Get(echo.HeaderContentType) {
	case echo.MIMEApplicationJSON:
		fallthrough
	case echo.MIMEApplicationJSONCharsetUTF8:
		err = ctx.JSONBlob(http.StatusOK, h.swaggerJSON)
	default:
		err = ctx.Blob(http.StatusOK, "application/yaml", openAPI)
	}

	if err != nil {
		return fmt.Errorf("could not write openapi definition into response: %w", err)
	}

	return nil
}

func NewOpenAPIHandler(openAPI *openapi3.T, logger *zerolog.Logger) *OpenAPIHandler {
	serversToLog := zerolog.Arr()
	for _, server := range openAPI.Servers {
		serversToLog = serversToLog.Str(server.URL)
	}

	log := logger.With().
		Str("api", openAPI.Info.Title).
		Str("version", openAPI.Info.Version).
		Array("servers", serversToLog).
		Logger()

	swaggerJSON, err := openAPI.MarshalJSON()
	if err != nil {
		log.Error().Err(err).Msg("could not marshal swagger json into memory")
	}

	log.Info().Msg("API Loaded!")

	return &OpenAPIHandler{openAPI, &log, swaggerJSON}
}
