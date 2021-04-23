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
	"AAA=",
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
