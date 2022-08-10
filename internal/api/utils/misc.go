package utils

import "context"

type apiVersionCtxKeyType string
type ApiVersion int

const (
	V1 ApiVersion = 1
	V2 ApiVersion = 2

	apiCtxKey apiVersionCtxKeyType = "api_version"
)

func (v ApiVersion) String() string {
	switch v {
	case V1:
		return "v1"
	case V2:
		return "v2"
	}
	return "unknown"
}

func WithApiVersion(ctx context.Context, value string) context.Context {
	return context.WithValue(ctx, apiCtxKey, value)
}

func GetApiVersion(ctx context.Context) string {
	result, ok := ctx.Value(apiCtxKey).(string)

	if !ok {
		return "unknown"
	}

	return result
}
