package controllers

import (
	"net/http"
	"playbook-dispatcher/internal/common/utils"

	"github.com/labstack/echo/v4"
)

type mockControllers struct {
}

func (this *mockControllers) ApiRunsList(ctx echo.Context, params ApiRunsListParams) error {
	return ctx.JSON(http.StatusOK, &Runs{
		Data: []Run{},
	})
}

func (this *mockControllers) ApiInternalRunsCreate(ctx echo.Context) error {
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
