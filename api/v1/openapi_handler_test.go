package v1_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	v1 "github.com/jakobmoellerdev/octi-sync-server/api/v1"
	"github.com/jakobmoellerdev/octi-sync-server/api/v1/REST"
)

func TestOpenAPIHandler_ServeOpenAPI(t *testing.T) {
	t.Parallel()
	log, assertions, api := SetupAPITest(t)
	swagger, err := REST.GetSwagger()

	assertions.NoError(err)

	handler := v1.NewOpenAPIHandler(swagger, &log)
	req := emptyRequest(http.MethodGet)
	rec := httptest.NewRecorder()
	ctx := api.NewContext(req, rec)

	for _, contentType := range []string{
		echo.MIMEApplicationJSONCharsetUTF8,
		echo.MIMEApplicationJSON,
		"some-random-content",
		"text/yaml",
	} {
		ctx.Reset(req, rec)
		req.Header.Set(echo.HeaderContentType, contentType)
		assertions.NoError(handler.ServeOpenAPI(ctx))
		assertions.NotEmpty(rec.Body)
	}
}
