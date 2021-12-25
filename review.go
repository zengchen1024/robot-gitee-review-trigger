package main

import (
	"time"

	sdk "github.com/opensourceways/go-gitee/gitee"
	"github.com/sirupsen/logrus"

	"github.com/opensourceways/robot-gitee-review-trigger/plugins"
	"github.com/opensourceways/robot-gitee-review-trigger/repoowners"
)

func (rt *robot) genRepoOwner(org, repo, branch string, cfg ownerConfig, log *logrus.Entry) (repoowners.RepoOwner, error) {
	if cfg.IsBranchWithoutOwners(branch) {
		cs, err := rt.client.listCollaborators(org, repo)
		if err != nil {
			return nil, err
		}
		return repoowners.RepoMemberAsOwners(cs), nil
	}

	return repoowners.NewRepoOwners(org, repo, branch, nil, log), nil
}

func (rt *robot) genPullRequest(prInfo iPRInfo, assignees []string, owner repoowners.RepoOwner) (pullRequest, error) {
	org, repo := prInfo.getOrgAndRepo()
	filenames, err := rt.client.getPullRequestChanges(org, repo, prInfo.getNumber())
	if err != nil {
		return pullRequest{}, err
	}

	return newPullRequest(prInfo, filenames, assignees, owner), nil
}

func (rt *robot) getReviewInfo(info iPRInfo) (ri reviewInfo, err error) {
	org, repo := info.getOrgAndRepo()

	ri.comments, err = rt.client.ListPRComments(org, repo, info.getNumber())
	if err != nil {
		return
	}

	ri.t, err = rt.client.getPRCodeUpdateTime(org, repo, info.getHeadSHA())
	return
}

type reviewInfo struct {
	comments []sdk.PullRequestComments
	t        time.Time
}

func (r reviewInfo) reviewGuides(botName string) []plugins.BotComment {
	return plugins.FindBotComment(r.comments, botName, isNotificationComment)
}

func (r reviewInfo) doStats(s *reviewStats, botName string) (reviewSummary, reviewResult) {
	return s.StatReview(r.comments, r.t, botName)
}
