// Package private provides primitives to interact the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen DO NOT EDIT.
package private

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
)

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Dispatch Playbooks
	// (POST /internal/dispatch)
	ApiInternalRunsCreate(ctx echo.Context) error
	// Cancel Playbook Runs
	// (POST /internal/v2/cancel)
	ApiInternalV2RunsCancel(ctx echo.Context) error
	// Obtain Connection Status of recipient(s) based on a list of host IDs
	// (POST /internal/v2/connection_status)
	ApiInternalHighlevelConnectionStatus(ctx echo.Context) error
	// Dispatch Playbooks
	// (POST /internal/v2/dispatch)
	ApiInternalV2RunsCreate(ctx echo.Context) error
	// Obtain connection status of recipient(s)
	// (POST /internal/v2/recipients/status)
	ApiInternalV2RecipientsStatus(ctx echo.Context) error
	// Get Version
	// (GET /internal/version)
	ApiInternalVersion(ctx echo.Context) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// ApiInternalRunsCreate converts echo context to params.
func (w *ServerInterfaceWrapper) ApiInternalRunsCreate(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.ApiInternalRunsCreate(ctx)
	return err
}

// ApiInternalV2RunsCancel converts echo context to params.
func (w *ServerInterfaceWrapper) ApiInternalV2RunsCancel(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.ApiInternalV2RunsCancel(ctx)
	return err
}

// ApiInternalHighlevelConnectionStatus converts echo context to params.
func (w *ServerInterfaceWrapper) ApiInternalHighlevelConnectionStatus(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.ApiInternalHighlevelConnectionStatus(ctx)
	return err
}

// ApiInternalV2RunsCreate converts echo context to params.
func (w *ServerInterfaceWrapper) ApiInternalV2RunsCreate(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.ApiInternalV2RunsCreate(ctx)
	return err
}

// ApiInternalV2RecipientsStatus converts echo context to params.
func (w *ServerInterfaceWrapper) ApiInternalV2RecipientsStatus(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.ApiInternalV2RecipientsStatus(ctx)
	return err
}

// ApiInternalVersion converts echo context to params.
func (w *ServerInterfaceWrapper) ApiInternalVersion(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.ApiInternalVersion(ctx)
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

	router.POST(baseURL+"/internal/dispatch", wrapper.ApiInternalRunsCreate)
	router.POST(baseURL+"/internal/v2/cancel", wrapper.ApiInternalV2RunsCancel)
	router.POST(baseURL+"/internal/v2/connection_status", wrapper.ApiInternalHighlevelConnectionStatus)
	router.POST(baseURL+"/internal/v2/dispatch", wrapper.ApiInternalV2RunsCreate)
	router.POST(baseURL+"/internal/v2/recipients/status", wrapper.ApiInternalV2RecipientsStatus)
	router.GET(baseURL+"/internal/version", wrapper.ApiInternalVersion)

}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"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",
	"PwEAAP//gWrQ4aUkAAA=",
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
