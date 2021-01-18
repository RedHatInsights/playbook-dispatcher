package api

import (
	"net/http"
	"playbook-dispatcher/utils"

	"github.com/labstack/echo/v4"
)

// TODO: this is a mock implementation
type serverInterfaceImpl struct {
}

func (this *serverInterfaceImpl) ApiRunsGet(ctx echo.Context, params ApiRunsGetParams) error {
	return ctx.JSON(http.StatusOK, &Runs{
		Data: []Run{},
	})
}

func (this *serverInterfaceImpl) ApiInternalRunsCreate(ctx echo.Context) error {
	var input []RunInput

	err := utils.ReadRequestBody(ctx, &input)
	if err != nil {
		// TODO: log
		return ctx.NoContent(http.StatusInternalServerError)
	}

	result := []*RunCreated{}

	for _, _ = range input {
		result = append(result, &RunCreated{
			Code: http.StatusNotImplemented,
		})
	}

	return ctx.JSON(http.StatusMultiStatus, result)
}
