package main

import (
	"github.com/opensourceways/community-robot-lib/giteeclient"
	"github.com/sirupsen/logrus"
)

//handleCanReviewComment handle the /can-review comment
func (bot *robot) handleCanReviewComment(e *noteEventInfo, cfg *botConfig, log *logrus.Entry) error {
	if !e.isCommentedByAuthor() {
		return nil
	}

	prInfo := prInfoOnNoteEvent{e.NoteEvent}

	if prInfo.hasLabel(labelCanReview) {
		return nil
	}

	if !prInfo.hasLabel(cfg.CLALabel) {
		tip := giteeclient.GenResponseWithReference(
			e.NoteEvent, "Please, sign cla first",
		)

		org, repo := prInfo.getOrgAndRepo()

		return bot.client.CreatePRComment(
			org, repo, prInfo.getNumber(), tip,
		)
	}

	if label := cfg.CI.LabelForBasicCIPassed; !cfg.CI.NoCI && label != "" && !prInfo.hasLabel(label) {
		tip := giteeclient.GenResponseWithReference(
			e.NoteEvent, "The basic CI should pass first",
		)

		org, repo := prInfo.getOrgAndRepo()

		return bot.client.CreatePRComment(
			org, repo, prInfo.getNumber(), tip,
		)
	}

	return bot.readyToReview(prInfo, cfg, log)
}
