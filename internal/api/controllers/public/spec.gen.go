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
