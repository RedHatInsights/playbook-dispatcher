package rbac

import (
	"context"
	"fmt"
	"net/http"
	"playbook-dispatcher/internal/common/constants"
	"regexp"
	"time"

	"github.com/redhatinsights/platform-go-middlewares/request_id"
	"github.com/spf13/viper"
)

const (
	applicationID  = "playbook-dispatcher"
	wildcard       = "*"
	operationEqual = "equal"
)

var permissionRegex = regexp.MustCompile(`^([[:ascii:]]+?):([[:ascii:]]+?):([[:ascii:]]+?)$`)

func NewRbacClientWithHttpRequestDoer(cfg *viper.Viper, doer HttpRequestDoer) RbacClient {
	client := &ClientWithResponses{
		ClientInterface: &Client{
			Server: cfg.GetString("rbac.host"),
			Client: doer,
			RequestEditor: func(ctx context.Context, req *http.Request) error {
				req.Header.Set(constants.HeaderRequestId, request_id.GetReqID(ctx))

				if identity, ok := ctx.Value(constants.HeaderIdentity).(string); ok {
					req.Header.Set(constants.HeaderIdentity, identity)
				}

				return nil
			},
		},
	}

	return &clientImpl{
		client: client,
	}
}

func NewRbacClient(cfg *viper.Viper) RbacClient {
	doer := http.Client{
		Timeout: time.Duration(cfg.GetInt64("rbac.timeout") * int64(time.Second)),
	}

	return NewRbacClientWithHttpRequestDoer(cfg, &doer)
}

type clientImpl struct {
	client ClientWithResponsesInterface
}

func matches(expected, actual string) bool {
	return actual == expected || actual == wildcard
}

/*
	playbook-dispatcher:run:read
	playbook-dispatcher:run:write
*/
func (this *clientImpl) GetPermissions(ctx context.Context) ([]Access, error) {
	res, err := this.client.GetPrincipalAccessWithResponse(ctx, &GetPrincipalAccessParams{
		Application: applicationID,
	})

	if err != nil {
		return nil, err
	}

	if res.JSON200 == nil {
		return nil, fmt.Errorf(`unexpected status code "%d" or content type "%s"`, res.HTTPResponse.StatusCode, res.HTTPResponse.Header.Get("content-type"))
	}

	if res.JSON200.Links.Next != nil {
		// should never happen but just in case
		return nil, fmt.Errorf(`RBAC page overflow, count: %d`, res.JSON200.Meta.Count)
	}

	return res.JSON200.Data, nil
}
