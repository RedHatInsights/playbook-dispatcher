package controllers

import dbModel "playbook-dispatcher/internal/common/model/db"

func dbRuntoApiRun(r *dbModel.Run) *Run {
	return &Run{
		Id:        RunId(r.ID.String()),
		Account:   Account(r.Account),
		Recipient: RunRecipient(r.Recipient.String()),
		Url:       Url(r.PlaybookURL),
		Timeout:   RunTimeout(r.Timeout),
		Status:    RunStatus(r.Status),
		Labels: Labels{
			AdditionalProperties: r.Labels,
		},
	}
}
