package main

import (
	"github.com/opensourceways/community-robot-lib/giteeclient"
	"github.com/sirupsen/logrus"
)

//handleCanReviewComment handle the /can-review comment
func (bot *robot) handleCanReviewComment(e *noteEventInfo, cfg *botConfig, log *logrus.Entry) error {
	if !e.isCommentedByPRAuthor() {
		return nil
	}

	prInfo := prInfoOnNoteEvent{e.NoteEvent}

	if prInfo.hasLabel(labelCanReview) {
		return nil
	}

	f := func(tip string) error {
		tip = giteeclient.GenResponseWithReference(e.NoteEvent, tip)
		org, repo := prInfo.getOrgAndRepo()

		return bot.client.CreatePRComment(
			org, repo, prInfo.getNumber(), tip,
		)
	}

	if !prInfo.hasLabel(cfg.CLALabel) {
		return f("Please, sign cla first")
	}

	if l := cfg.LabelsForBasicCIPassed; len(l) > 0 && !prInfo.hasAnyLabel(l) {
		return f("The basic CI should pass first")
	}

	return bot.readyToReview(prInfo, cfg, log)
}
