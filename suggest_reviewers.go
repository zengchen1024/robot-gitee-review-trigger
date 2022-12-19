package main

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"net/http"
	"sort"

	"github.com/opensourceways/community-robot-lib/utils"
	"github.com/opensourceways/repo-owners-cache/repoowners"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/sets"
)

func suggestReviewers(
	c ghclient, owner repoowners.RepoOwner, pr iPRInfo,
	reviewerCount int, endpoint string, log *logrus.Entry,
) ([]string, error) {
	v, err := findReviewers(c, owner, pr, reviewerCount, log)
	if err != nil {
		return nil, err
	}

	if endpoint != "" {
		return recommendReviewers(endpoint, v, pr)
	}

	if len(v) <= reviewerCount {
		return v, nil
	}

	return selectReviewer(v, reviewerCount), nil
}

func findReviewers(
	c ghclient, owner repoowners.RepoOwner,
	pr iPRInfo, reviewerCount int, log *logrus.Entry,
) ([]string, error) {
	org, repo := pr.getOrgAndRepo()
	changes, err := c.getPullRequestChanges(org, repo, pr.getNumber())
	if err != nil {
		return nil, err
	}

	excludedReviewers := sets.NewString(normalizeLogin(pr.getAuthor()))

	reviewers := retrieveReviewers(owner, changes, reviewerCount, excludedReviewers)
	if len(reviewers) < reviewerCount {

		approvers := retrieveReviewers(
			fallbackReviewersClient{oc: owner},
			changes,
			reviewerCount-len(reviewers),
			excludedReviewers.Insert(reviewers...),
		)
		reviewers = append(reviewers, approvers...)
		sort.Strings(reviewers)

		log.Infof("Added %d approvers as reviewers.", len(approvers))
	}

	if n := len(reviewers); n < reviewerCount {
		log.Warnf(
			"Not enough reviewers found in OWNERS files for files touched by this PR. %d/%d reviewers found.",
			n, reviewerCount,
		)
	}

	return reviewers, nil
}

func retrieveReviewers(
	rc reviewersClient, files []string,
	minReviewers int, excludedReviewers sets.String,
) []string {
	// leaf first
	leafReviewers := sets.NewString()

	for _, filename := range files {
		if v := rc.LeafReviewers(filename); v.Len() > 0 {
			leafReviewers = leafReviewers.Union(v)
		}
	}

	if excludedReviewers.Len() > 0 {
		leafReviewers = leafReviewers.Difference(excludedReviewers)
	}

	if leafReviewers.Len() >= minReviewers {
		return leafReviewers.UnsortedList()
	}

	// all reviewers
	fileReviewers := sets.NewString()

	for _, filename := range files {
		if v := rc.Reviewers(filename); v.Len() > 0 {
			fileReviewers = fileReviewers.Union(v)
		}
	}

	if excludedReviewers.Len() > 0 {
		fileReviewers = fileReviewers.Difference(excludedReviewers)
	}

	return fileReviewers.UnsortedList()
}

func selectReviewer(list []string, n int) []string {
	sort.Strings(list)

	ln := len(list)
	if ln <= n || n <= 0 {
		return list
	}

	for i := 0; i < n; i++ {
		j := rand.Intn(ln - i)
		k := ln - i - 1
		list[j], list[k] = list[k], list[j]
	}
	return list[ln-n:]
}

type reviewersClient interface {
	Reviewers(path string) sets.String
	LeafReviewers(path string) sets.String
}

type fallbackReviewersClient struct {
	oc repoowners.RepoOwner
}

func (foc fallbackReviewersClient) Reviewers(path string) sets.String {
	return foc.oc.Approvers(path)
}

func (foc fallbackReviewersClient) LeafReviewers(path string) sets.String {
	return foc.oc.LeafApprovers(path)
}

type reviewerRecommendReq struct {
	Community string   `json:"community"`
	PrUrl     string   `json:"prUrl"`
	PrTitle   string   `json:"prTitle"`
	Reviewers []string `json:"reviewers"`
}

type reviewerRecommendResp struct {
	Msg  string   `json:"msg"`
	Code int      `json:"code"`
	Data []string `json:"data"`
}

func recommendReviewers(endpoint string, reviewers []string, pr iPRInfo) ([]string, error) {
	org, _ := pr.getOrgAndRepo()
	payload, err := json.Marshal(reviewerRecommendReq{
		Community: org,
		PrUrl:     pr.getUrl(),
		PrTitle:   pr.getTitle(),
		Reviewers: reviewers,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "robot-gitee-review-trigger")

	res := new(reviewerRecommendResp)
	cli := utils.HttpClient{MaxRetries: 3}
	if err = cli.ForwardTo(req, res); err != nil {
		return nil, err
	}

	return res.Data, nil
}
