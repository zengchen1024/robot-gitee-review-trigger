package main

import (
	"fmt"
	"strings"

	"github.com/opensourceways/community-robot-lib/giteeclient"
	sdk "github.com/opensourceways/go-gitee/gitee"
	"github.com/sirupsen/logrus"
)

func (rt *robot) handleNoteEvent1(e *sdk.NoteEvent, cfg *botConfig, log *logrus.Entry) error {
	if !e.IsPullRequest() || !e.IsPROpen() {
		return nil
	}

	if e.IsCreatingCommentEvent() && e.GetCommenter() != rt.botName {
		if cmds := parseReviewCommand(e.GetComment().GetBody()); len(cmds) > 0 {
			return rt.handleReviewComment(e, cfg, log)
		}
	}

	return rt.handleCIStatusComment(e, cfg, log)
}

func (rt *robot) handleReviewComment(e *sdk.NoteEvent, cfg *botConfig, log *logrus.Entry) error {
	org, repo := e.GetOrgRepo()
	owner, err := rt.genRepoOwner(org, repo, e.GetPRBaseRef(), cfg.Owner, log)
	if err != nil {
		return err
	}

	prInfo := prInfoOnNoteEvent{e}
	pr, err := rt.genPullRequest(prInfo, getAssignees(e.GetPullRequest()), owner)
	if err != nil {
		return err
	}

	stats := &reviewStats{
		pr:        &pr,
		cfg:       cfg.Review,
		reviewers: owner.AllReviewers(),
	}

	cmd, validReview := rt.isValidReview(stats, e, log)
	if !validReview {
		return nil
	}

	info, err := rt.getReviewInfo(prInfo)
	if err != nil {
		return err
	}

	canReview := cfg.CI.NoCI || stats.pr.info.hasLabel(cfg.CI.LabelForCIPassed)
	pa := PostAction{
		c:                rt.client,
		cfg:              cfg,
		owner:            owner,
		log:              log,
		pr:               &pr,
		isStartingReview: canReview,
	}

	oldTips := info.reviewGuides(rt.botName)
	rs, rr := info.doStats(stats, rt.botName)
	return pa.do(oldTips, cmd, rs, rr, rt.botName)
}

func (rt *robot) isValidReview(stats *reviewStats, e *sdk.NoteEvent, log *logrus.Entry) (string, bool) {
	commenter := normalizeLogin(e.GetCommenter())

	cmd, invalidCmd := getReviewCommand(e.GetComment().GetBody(), commenter, stats.genCheckCmdFunc())

	validReview := cmd != "" && stats.isReviewer(commenter)

	if !validReview {
		log.Infof(
			"It can't handle note event, because cmd(%s) is empty or commenter(%s) is not a reviewer. There are %d reviewers.",
			cmd, commenter, stats.numberOfReviewers(),
		)
	}

	if invalidCmd != "" {

		info := stats.pr.info
		org, repo := info.getOrgAndRepo()

		s := fmt.Sprintf(
			"You can't comment `/%s`. Please see the [*Command Usage*](%s) to get detail.",
			strings.ToLower(invalidCmd),
			"",
		)

		rt.client.CreatePRComment(
			org, repo, info.getNumber(),
			giteeclient.GenResponseWithReference(e, s),
		)
	}

	return cmd, validReview
}
