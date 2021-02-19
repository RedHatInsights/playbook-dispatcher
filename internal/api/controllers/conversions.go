package controllers

import (
	dbModel "playbook-dispatcher/internal/common/model/db"
	"playbook-dispatcher/internal/common/utils"

	"github.com/google/uuid"
)

const (
	fieldId        = "id"
	fieldAccount   = "account"
	fieldRecipient = "recipient"
	fieldUrl       = "url"
	fieldLabels    = "labels"
	fieldTimeout   = "timeout"
	fieldStatus    = "status"
	fieldCreatedAt = "created_at"
	fieldUpdatedAt = "updated_at"
	fieldRun       = "run"
	fieldHost      = "host"
	fieldStdout    = "stdout"
)

var (
	runFields     = utils.IndexStrings(fieldId, fieldAccount, fieldRecipient, fieldUrl, fieldLabels, fieldTimeout, fieldStatus, fieldCreatedAt, fieldUpdatedAt)
	runHostFields = utils.IndexStrings(fieldHost, fieldRun, fieldStatus, fieldStdout)
)

var defaultRunFields = []string{
	fieldId,
	fieldRecipient,
	fieldUrl,
	fieldLabels,
	fieldTimeout,
	fieldStatus,
}

var defaultRunHostFields = []string{
	fieldHost,
	fieldRun,
	fieldStatus,
}

func dbRuntoApiRun(r *dbModel.Run, fields []string) *Run {
	run := Run{}

	for _, field := range fields {
		switch field {
		case fieldId:
			run.Id = (*RunId)(convertUuid(r.ID))
		case fieldAccount:
			value := Account(r.Account)
			run.Account = &value
		case fieldRecipient:
			run.Recipient = (*RunRecipient)(convertUuid(r.Recipient))
		case fieldUrl:
			value := Url(r.URL)
			run.Url = &value
		case fieldLabels:
			run.Labels = (&Labels{
				AdditionalProperties: r.Labels,
			})
		case fieldTimeout:
			value := RunTimeout(r.Timeout)
			run.Timeout = &value
		case fieldStatus:
			value := RunStatus(r.Status)
			run.Status = &value
		case fieldCreatedAt:
			val := CreatedAt(r.CreatedAt)
			run.CreatedAt = &val
		case fieldUpdatedAt:
			val := UpdatedAt(r.UpdatedAt)
			run.UpdatedAt = &val
		default:
			panic("unknown field " + field)
		}
	}

	return &run
}

func convertUuid(value uuid.UUID) *string {
	result := value.String()
	return &result
}
