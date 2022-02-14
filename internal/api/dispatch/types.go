package dispatch

import (
	"context"
	"fmt"
	"playbook-dispatcher/internal/common/model/generic"

	"github.com/google/uuid"
)

// orchestrates sending of playbook run signal and storing the database records
type DispatchManager interface {
	ProcessRun(ctx context.Context, account string, service string, run generic.RunInput) (runID, correlationID uuid.UUID, err error)
}

// Indicates that the recipient is not connected
type RecipientNotFoundError struct {
	err       error
	recipient uuid.UUID
}

func (this *RecipientNotFoundError) Error() string {
	return fmt.Sprintf("Recipient not found: %s", this.recipient)
}
