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
	router.POST(baseURL+"/internal/v2/dispatch", wrapper.ApiInternalV2RunsCreate)
	router.POST(baseURL+"/internal/v2/recipients/status", wrapper.ApiInternalV2RecipientsStatus)
	router.GET(baseURL+"/internal/version", wrapper.ApiInternalVersion)

}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"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",
	"yFzX5dv32z8CAAD//4VkCVUgFwAA",
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
