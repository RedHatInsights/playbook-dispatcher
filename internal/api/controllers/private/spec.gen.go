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
	"//8834Lj3iUAAA==",
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
