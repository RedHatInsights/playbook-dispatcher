package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/ghodss/yaml"
	"github.com/labstack/echo/v4"
	"github.com/qri-io/jsonschema"
	"github.com/spf13/viper"
)

func DieOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func StopServer(ctx context.Context, server *echo.Echo) {
	if e := server.Shutdown(ctx); e != nil {
		GetLogFromContext(ctx).Error(e)
	}
}

func IndexStrings(values ...string) map[string]string {
	result := make(map[string]string, len(values))

	for _, value := range values {
		result[value] = value
	}

	return result
}

func MapStrings(values []string, fn func(string) string) []string {
	result := make([]string, len(values))

	for i, value := range values {
		result[i] = fn(value)
	}

	return result
}

func WgWaitFor(wg *sync.WaitGroup, timeout time.Duration) error {
	channel := make(chan struct{})

	go func() {
		wg.Wait()
		close(channel)
	}()

	select {
	case <-channel:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("Time out waiting for WaitGroup")
	}
}

func BuildUrl(base string, keysAndValues ...interface{}) string {
	if len(keysAndValues)%2 != 0 {
		panic(fmt.Sprintf("Odd number of parameters: %s", keysAndValues))
	}

	params := make([]string, len(keysAndValues)/2)
	for i := 0; i < len(keysAndValues)/2; i++ {
		params[i] = fmt.Sprintf("%s=%s",
			url.QueryEscape(fmt.Sprintf("%s", keysAndValues[i*2])),
			url.QueryEscape(fmt.Sprintf("%v", keysAndValues[(i*2)+1])),
		)
	}

	return fmt.Sprintf("%s?%s", base, strings.Join(params, "&"))
}

func MustMarshal(value interface{}) []byte {
	bytes, err := json.Marshal(value)
	DieOnError(err)
	return bytes
}

func MustUnmarshal(bytes []byte, result interface{}) {
	err := json.Unmarshal(bytes, result)
	DieOnError(err)
}

func SetRequestContextValue(c echo.Context, key interface{}, value interface{}) echo.Context {
	req := c.Request()
	c.SetRequest(req.WithContext(context.WithValue(req.Context(), key, value)))
	return c
}

func LoadSchemas(cfg *viper.Viper, schemaNames []string) (schemas []*jsonschema.Schema) {
	for _, schemaName := range schemaNames {
		var schema jsonschema.Schema
		file, err := ioutil.ReadFile(cfg.GetString(schemaName))
		DieOnError(err)
		err = yaml.Unmarshal(file, &schema)
		DieOnError(err)

		schemas = append(schemas, &schema)
	}
	return
}

func IsOrgIdBlocklisted(cfg *viper.Viper, orgId string) bool {
	blocklistedOrgIds := strings.Split(cfg.GetString("blocklist.orgids"), ",")

	if len(blocklistedOrgIds) > 0 {
		for _, blockedOrgId := range blocklistedOrgIds {
			if blockedOrgId == orgId {
				return true
			}
		}
	}

	return false
}
