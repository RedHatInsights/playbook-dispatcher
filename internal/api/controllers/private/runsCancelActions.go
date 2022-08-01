package private

import (
	"net/http"
	"playbook-dispatcher/internal/api/controllers/public"
	"playbook-dispatcher/internal/api/dispatch"
	"playbook-dispatcher/internal/common/model/generic"

	"github.com/google/uuid"
)

func CancelInputV2GenericMap(
	cancelInput CancelInputV2,
	runId uuid.UUID,
) generic.CancelInput {
	orgIdString := string(cancelInput.OrgId)
	principal := string(cancelInput.Principal)

	result := generic.CancelInput{
		RunId:     runId,
		OrgId:     orgIdString,
		Principal: principal,
	}

	return result
}

func runCancelError(code int) *RunCanceled {
	return &RunCanceled{
		Code: code,
		// TODO report error
	}
}

func handleRunCancelError(err error) *RunCanceled {
	if _, ok := err.(*dispatch.RunNotFoundError); ok {
		return runCancelError(http.StatusNotFound)
	}

	if _, ok := err.(*dispatch.RunOrgIdMismatchError); ok {
		return runCancelError(http.StatusBadRequest)
	}

	if _, ok := err.(*dispatch.RecipientNotFoundError); ok {
		return runCancelError(http.StatusConflict)
	}

	if _, ok := err.(*dispatch.RunCancelNotCancelableError); ok {
		return runCancelError(http.StatusConflict)
	}

	if _, ok := err.(*dispatch.RunCancelTypeError); ok {
		return runCancelError(http.StatusBadRequest)
	}

	return runCancelError(http.StatusInternalServerError)
}

func runCanceled(runID uuid.UUID) *RunCanceled {
	return &RunCanceled{
		Code:  http.StatusAccepted,
		RunId: public.RunId(runID.String()),
	}
}
