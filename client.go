package main

import (
	"strings"
	"time"

	sdk "github.com/opensourceways/go-gitee/gitee"
)

type ghclient struct {
	iClient
}

func (c ghclient) getPRCodeUpdateTime(org, repo, headSHA string) (time.Time, error) {
	v, err := c.GetPRCommit(org, repo, headSHA)
	if err != nil {
		return time.Time{}, err
	}

	return v.Commit.Committer.Date, nil
}

func (c ghclient) getPullRequestChanges(org, repo string, number int32) ([]string, error) {
	filenames, err := c.GetPullRequestChanges(org, repo, number)
	if err != nil {
		return nil, err
	}

	r := make([]string, 0, len(filenames))
	for i := range filenames {
		r = append(r, filenames[i].Filename)
	}
	return r, nil
}

func (c ghclient) listCollaborators(org, repo string) ([]string, error) {
	cs, err := c.ListCollaborators(org, repo)
	if err != nil {
		return nil, err
	}

	r := make([]string, 0, len(cs))
	for i := range cs {
		r = append(r, normalizeLogin(cs[i].Login))
	}
	return r, nil
}

func getAssignees(pr *sdk.PullRequestHook) []string {
	if pr == nil {
		return nil
	}

	v := pr.Assignees
	as := make([]string, 0, len(v))
	for i := range v {
		as = append(as, normalizeLogin(v[i].Login))
	}
	return as
}

func normalizeLogin(s string) string {
	return strings.TrimPrefix(strings.ToLower(s), "@")
}
