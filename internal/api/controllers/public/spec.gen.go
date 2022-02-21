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

	"H4sIAAAAAAAC/9RZ3W8bxxH/VxbbPtIkZTlByqdKSoMKVaxAjoAAqWAvb4fk2ne75/2gxBr834uZve87",
	"6kglLew33nFmdnY+fzP3hScmy40G7R1ffOG5sCIDD5aeblSmPP6Q4BKrcq+M5gv+s3hSWciYDtkSLDMr",
	"ZsGF1DvmDbPgg9V8whWSfg5gd3zCtciAL3hKAifcJRvIRJS8EiH1fPHdfMKzKJgvXs/xSen4dDbhfpcj",
	"v9Ie1mD5fj/ht6uVgwHtrrVUifDgmN8Ac15Yr/Sa5cYppEB18Q/SjFlIhVdbQM3xLVojBQ/MgUdK5SFD",
	"QcKzTPhkU7MeuKGJWg1esXmn+eCd7oL+p3H+JwWpdP2r/QgrpcGxFf2POi+hMDhIpjRpZ8HlRjuY/hu9",
	"AE95aiTwhbcBhlWO0loq59bkYL2CqITw7Yv8zjfG0SW98AFZbdD8YcLJXEgKGi9Z0eHfDWrnpQn4PlX6",
	"kyNLbkF7Y3fvlUQ5hWmct0qv+b56IawVO7JU8cIsP0LikcL5XYpvJEB+W72tDJp6sH2DXqSpeXRsZSxb",
	"EQlGylI4kMxothVWmeBYYhX+JY41J5112Jytyy6+8L9aWPEF/8usTsVZ5HWz65L2Wr4NaSqWKeC10J6L",
	"L1yXrwp1OueQ9JWxmfB8wUNQErXs8PRMnYolpG5Msbugb4iwqZYDu1UJjPG+i2Q157A/KVjGRBHVmKQD",
	"keG+/jwjl4kkMUFTHkGicgX0O9iUV96acK8yiElVWG4oHZ8Rlxgbi6HR74lsTH7t7glPLAgP8r0gxXJZ",
	"PvzJqey+qjz+A4lSG/4lGfrSNPv/JNU7Y/3lru8jfM+MlWTTIYM7Y/375a5l8TJwWwFWPyyES9ov8MyB",
	"sNuT0WPC0v0vhbyDzwEcuSAx2hfeEHmeIn5QRs8+OkN1ttbnObP9w1pj41Htm18KycrD9hP+k7FLJSXo",
	"//3JF0kCzpXgZq22oLFumWATYMoxbTwTmDsgyc2FQDzvoqgSfXwlQXu1UhH5oVwPWlANycTTDei13/DF",
	"WQQ71eNAGF9Fr10MnHHBsNw4L7KcPW4gFlzQ3u7Yo8BEJk4+qXMHS84rZOIDJ0X79BI4A+fEmvKoGy4Y",
	"LZ+DsiAx+krCh4FEGOrQL0nqm6qaCCkJqYr0l5a6PZaOzSo2loEX2EyYWJrgyXi/pGK3NOYTs0FP2ZXQ",
	"2NICFsh20cyDzY0DN+UDd70huHZQxZVIXQ+FrJR1Ax6ugDdCwDJAiZblYg1dlE7jBR9EK0dLR9LThGt4",
	"OlY4kp4mPLewxbZ05AEl+SmHdMI4uqKw2VAs/wwRgDzr3u60FSMdJ6sYbRVOwjJBnN2QOFBY+uNkKYoa",
	"tsDJLM6J3clpwr3xIu2LpNcDcyrNchjuzXGuOuLs7M3gdNa0ZQmh4sFDxry162t5UvmsFODfnZ/98Ppv",
	"85NL6l2cC9rmFnUlf66JlAV/38OC4wjnqma4liSh7tcj3HUX2E/4UYfFE45DX0VV3U+4sesj7hKd1kVo",
	"IwrdVbQng7PjQdld0BGXUbgXYHyc59eCct/C5SN895Ey+gTngDF6mw6jxF5o9NLhXqvPAZiqs4KakjfF",
	"vuXR2E/MRvDEHpXfsBrGDScAzvv9JNiYoT7xVmRQHVjosGOCITWdpjS70E4tU2DV2D50bnem77b/ft8q",
	"W+mI//A2se3WY/8IywtjqljK9Gx0G3wePMutkSEByZY7xBAaq2dpmgpcGN0AmsXup9+ThsKkvucz+5LS",
	"iSNg6pkj3OGpuxqWj/BIf4490qOVK7Oi1T5HTO2423hI3YK/PPRh+MLH5ZtZMdGCh01YfSh8+1PtKcD1",
	"gP9ayt81C/BYC6V89YY9blSyYaIIwOpSyjEhpQXn2kPDM7d7VyVQ++yrYC1oz2KCDRqvnFmLLMF5NtAc",
	"hkcLlQYLrX1KInQCKQwvPhslvLkZOv9+Pu8NABn2b1TJQWK0dEysPNjCKLSyChptkRjtlASL4F+oFCST",
	"IW7AK52qTfz38zc/zEcW13Hu/+OZ9Q1k1bu6u3e2G/GP+KnAW7Vek33ratgJkhEg193d4LTa4hidJztL",
	"nMYu5SVxOXpcDRlOHeZpLCuwydET/b0dQPv3dzdUCUpgX5q8lfK01RzYDim9MuVCRiR0DciESvmCfzT/",
	"gdXfLciN8NPEZP3Zu3Luj8rliFvAUtFhBRqkIftQ13TYNlFvykwNiQfJtkqwq9QEya7iO2NpJFeeBoSB",
	"A/mEb8G6qNDZdD6dE+zNQYtc8QU/n86n5ziLCb+hjJqJXM3yQtArWQmabc9mNmjqtkS4HvrMdkezGd4x",
	"VY6qTswSnNpyCw5LtF4X91J6a9Jt3J03E8FN2b1OwSETOoOuHhwyxpWEK1fttLJyzOUWhGQiscY5loXU",
	"qzyFrsy3hmVg1yjGWCZBhmqzhm7JwWI8RCzjN8rVO7FXTE1hytSqBIG/IUNT/bx5FLtgQkt2iVpq5h8N",
	"c2FZa0t4FZ6U8xNmNLQt81sdECQECTBMLuMHBSykFWrmF7kqEcyNIljV/FL7+3DFq0lm7e9g+8nxDPTN",
	"4giG+K34CMLiu+3+obORfT2f/2kL0QrtDexEb/+FefEmnjYkpNJq1lgSE8v5OEu93KV9asgyYXd8wdFr",
	"Y8lALCNZeUpCtoRTbNXtqFgKTNmvG4gMSdwJxryr0hA5PsR3H1jlxchWfB7rfxgrojwmXCEX/WpNmtKE",
	"h5I/RPam1IOB/+KgdydFvDs+3BtfOr7B5PjaEqObBsUepfQzrSD4xvvcLWazBNvitNGOuxeJoyoFX5RC",
	"lo9CZj3qatuKjZHvH/b/DQAA///ugd+HFCMAAA==",
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
