// Package public provides primitives to interact the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen DO NOT EDIT.
package public

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/deepmap/oapi-codegen/pkg/runtime"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
)

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// List hosts involved in Playbook runs
	// (GET /api/playbook-dispatcher/v1/run_hosts)
	ApiRunHostsList(ctx echo.Context, params ApiRunHostsListParams) error
	// List Playbook runs
	// (GET /api/playbook-dispatcher/v1/runs)
	ApiRunsList(ctx echo.Context, params ApiRunsListParams) error
	// Get API version
	// (GET /internal/version)
	ApiVersion(ctx echo.Context, params ApiVersionParams) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// ApiRunHostsList converts echo context to params.
func (w *ServerInterfaceWrapper) ApiRunHostsList(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params ApiRunHostsListParams
	// ------------- Optional query parameter "filter" -------------

	err = runtime.BindQueryParameter("deepObject", true, false, "filter", ctx.QueryParams(), &params.Filter)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter filter: %s", err))
	}

	// ------------- Optional query parameter "fields" -------------

	err = runtime.BindQueryParameter("deepObject", true, false, "fields", ctx.QueryParams(), &params.Fields)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter fields: %s", err))
	}

	// ------------- Optional query parameter "limit" -------------

	err = runtime.BindQueryParameter("form", true, false, "limit", ctx.QueryParams(), &params.Limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter limit: %s", err))
	}

	// ------------- Optional query parameter "offset" -------------

	err = runtime.BindQueryParameter("form", true, false, "offset", ctx.QueryParams(), &params.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter offset: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.ApiRunHostsList(ctx, params)
	return err
}

// ApiRunsList converts echo context to params.
func (w *ServerInterfaceWrapper) ApiRunsList(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params ApiRunsListParams
	// ------------- Optional query parameter "filter" -------------

	err = runtime.BindQueryParameter("deepObject", true, false, "filter", ctx.QueryParams(), &params.Filter)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter filter: %s", err))
	}

	// ------------- Optional query parameter "fields" -------------

	err = runtime.BindQueryParameter("deepObject", true, false, "fields", ctx.QueryParams(), &params.Fields)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter fields: %s", err))
	}

	// ------------- Optional query parameter "sort_by" -------------

	err = runtime.BindQueryParameter("form", true, false, "sort_by", ctx.QueryParams(), &params.SortBy)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter sort_by: %s", err))
	}

	// ------------- Optional query parameter "limit" -------------

	err = runtime.BindQueryParameter("form", true, false, "limit", ctx.QueryParams(), &params.Limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter limit: %s", err))
	}

	// ------------- Optional query parameter "offset" -------------

	err = runtime.BindQueryParameter("form", true, false, "offset", ctx.QueryParams(), &params.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter offset: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.ApiRunsList(ctx, params)
	return err
}

func (w *ServerInterfaceWrapper) ApiVersion(ctx echo.Context) error {
	var params ApiVersionParams
	err := w.Handler.ApiVersion(ctx, params)
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

	router.GET(baseURL+"/api/playbook-dispatcher/v1/run_hosts", wrapper.ApiRunHostsList)
	router.GET(baseURL+"/api/playbook-dispatcher/v1/runs", wrapper.ApiRunsList)
	router.GET(baseURL+"/internal/version", wrapper.ApiVersion)

}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"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",
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file.
func GetSwagger() (*openapi3.Swagger, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %s", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}

	swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("error loading Swagger: %s", err)
	}
	return swagger, nil
}
