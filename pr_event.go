package main

import (
	"fmt"
	"strings"

	sdk "github.com/opensourceways/go-gitee/gitee"
	"github.com/sirupsen/logrus"

	"github.com/opensourceways/robot-gitee-review-trigger/plugins"
)

type prInfoOnPREvent struct {
	e *sdk.PullRequestEvent
}

func (pr prInfoOnPREvent) getOrgAndRepo() (string, string) {
	return pr.e.GetOrgRepo()
}

func (pr prInfoOnPREvent) getNumber() int32 {
	return pr.e.GetPRNumber()
}

func (pr prInfoOnPREvent) getTargetBranch() string {
	return pr.e.GetPRBaseRef()
}

func (pr prInfoOnPREvent) hasLabel(l string) bool {
	return pr.e.GetPRLabelSet().Has(l)
}
func (pr prInfoOnPREvent) getAuthor() string {
	return pr.e.GetPRAuthor()
}

func (pr prInfoOnPREvent) getHeadSHA() string {
	return pr.e.GetPRHeadSha()
}

func (rt *robot) handlePREvent1(e *sdk.PullRequestEvent, cfg *botConfig, log *logrus.Entry) error {
	canReview := cfg.CI.NoCI

	switch sdk.GetPullRequestAction(e) {
	case sdk.ActionOpen:
		mr := multiError()
		pr := prInfoOnPREvent{e}

		if cfg.NeedWelcome {
			if err := rt.welcome(pr, cfg.Doc); err != nil {
				mr.Add(fmt.Sprintf("add welcome comment, err:%s", err.Error()))
			}
		}

		if canReview {
			if err := rt.readyToReview(pr, cfg, log); err != nil {
				mr.AddError(err)
			}
		}
		return mr.Err()

	case sdk.PRActionChangedSourceBranch:
		toKeep := []string{}
		if canReview {
			toKeep = append(toKeep, labelCanReview)
		}
		return rt.resetToReview(prInfoOnPREvent{e}, cfg, toKeep, log)
	}

	return nil
}

func (rt *robot) welcome(pr iPRInfo, doc string) error {
	org, repo := pr.getOrgAndRepo()

	return rt.client.CreatePRComment(
		org, repo, pr.getNumber(),
		fmt.Sprintf(
			`
Thank your for your pull-request.

The full list of commands accepted by me can be found at [**here**](%s).

%s
`,
			"",
			doc,
		),
	)
}

func (rt *robot) readyToReview(pr iPRInfo, cfg *botConfig, log *logrus.Entry) error {
	mr := multiError()

	if err := rt.addLabelOfCanReview(pr); err != nil {
		mr.AddError(err)
	}

	if err := rt.addReviewNotification(pr, cfg, log); err != nil {
		mr.AddError(err)
	}

	return mr.Err()
}

func (rt *robot) addLabelOfCanReview(pr iPRInfo) error {
	l := labelCanReview
	if pr.hasLabel(l) {
		return nil
	}

	org, repo := pr.getOrgAndRepo()
	return rt.client.AddPRLabel(org, repo, pr.getNumber(), l)
}

func (rt *robot) addReviewNotification(pr iPRInfo, cfg *botConfig, log *logrus.Entry) error {
	org, repo := pr.getOrgAndRepo()
	owner, err := rt.genRepoOwner(org, repo, pr.getTargetBranch(), cfg.Owner, log)
	if err != nil {
		return err
	}

	reviewers, err := suggestReviewers(rt.client, owner, pr, cfg.Review.TotalNumberOfReviewers, log)
	if err != nil {
		return fmt.Errorf("suggest reviewers, err: %s", err.Error())
	}

	if len(reviewers) == 0 {
		return nil
	}

	s := newNotificationComment(&reviewSummary{}, "", rt.botName).startReviewComment(reviewers)

	return rt.client.CreatePRComment(org, repo, pr.getNumber(), s)
}

func (rt *robot) resetToReview(pr iPRInfo, cfg *botConfig, toKeep []string, log *logrus.Entry) error {
	mr := multiError()

	if err := rt.resetLabels(pr, cfg, toKeep); err != nil {
		mr.Add(fmt.Sprintf("remove label when source code changed, err:%s", err.Error()))
	}

	if err := rt.deleteReviewNotifiction(pr); err != nil {
		mr.Add(fmt.Sprintf("delete tips, err:%s", err.Error()))
	}

	if err := rt.addReviewNotification(pr, cfg, log); err != nil {
		mr.AddError(err)
	}

	return mr.Err()
}

func (rt *robot) resetLabels(pr iPRInfo, cfg *botConfig, toKeep []string) error {
	rmls, err := updateAndReturnRemovedLabels(rt.client, pr, toKeep...)
	if err != nil {
		return err
	}

	if len(rmls) > 0 {
		org, repo := pr.getOrgAndRepo()

		rt.client.CreatePRComment(
			org, repo, pr.getNumber(), fmt.Sprintf(
				"New changes are detected. Remove the following labels: %s.",
				strings.Join(rmls, ", "),
			),
		)
	}

	return nil
}

func (rt *robot) deleteReviewNotifiction(pr iPRInfo) error {
	org, repo := pr.getOrgAndRepo()

	comments, err := rt.client.ListPRComments(org, repo, pr.getNumber())
	if err != nil {
		return err
	}

	cs := plugins.FindBotComment(comments, rt.botName, isNotificationComment)
	for _, c := range cs {
		rt.client.DeletePRComment(org, repo, c.CommentID)
	}
	return nil
}
