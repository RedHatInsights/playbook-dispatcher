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
