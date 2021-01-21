package controllers

import "playbook-dispatcher/models"

func dbRuntoApiRun(r *models.Run) *Run {
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
