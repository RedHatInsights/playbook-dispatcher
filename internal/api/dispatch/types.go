package dispatch

import (
	"context"
	"fmt"
	"playbook-dispatcher/internal/common/model/generic"

	"github.com/google/uuid"
)

// orchestrates sending of playbook run signal and storing the database records
type DispatchManager interface {
	ProcessRun(ctx context.Context, account string, service string, run generic.RunInput, api_verison string) (runID, correlationID uuid.UUID, err error)
	ProcessCancel(ctx context.Context, account string, cancel generic.CancelInput) (runID, correlationID uuid.UUID, err error)
}

// Indicates that the recipient is not connected
type RecipientNotFoundError struct {
	err       error
	recipient uuid.UUID
}

type RunNotFoundError struct {
	err   error
	runID uuid.UUID
}

type RunCancelTypeError struct {
	err   error
	runID uuid.UUID
}
type RunCancelNotCancelableError struct {
	runID uuid.UUID
}

func (this *RecipientNotFoundError) Error() string {
	return fmt.Sprintf("Recipient not found: %s", this.recipient)
}

func (this *RunNotFoundError) Error() string {
	return fmt.Sprintf("Run not found: %s", this.runID)
}

func (this *RunCancelTypeError) Error() string {
	return fmt.Sprintf("Run not of type RHC Satellite and cannot be canceled: %s", this.runID)
}

func (this *RunCancelNotCancelableError) Error() string {
	return fmt.Sprintf("Run has finished running and cannot be canceled: %s", this.runID)
}
