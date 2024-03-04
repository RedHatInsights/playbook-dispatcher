package utils

import (
	"encoding/json"
	"io/ioutil"

	echo "github.com/labstack/echo/v4"
)

// workaround for https://github.com/labstack/echo/issues/1356
func ReadRequestBody(ctx echo.Context, i interface{}) error {
	body, err := ioutil.ReadAll(ctx.Request().Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, i)
	if err != nil {
		return err
	}

	return nil
}
