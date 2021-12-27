package main

import (
	"github.com/opensourceways/repo-owners-cache/repoowners"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/sets"
)

type suggestingApprover struct {
	pr    *pullRequest
	cfg   reviewConfig
	owner repoowners.RepoOwner
}

func (p suggestingApprover) selectApprovers(as []string, n int) []string {
	excluded := sets.NewString(as...)

	if !p.cfg.AllowSelfApprove {
		excluded.Delete(p.pr.prAuthor())
	}

	return getReviewers(fakeReviewersClient{s: &p}, p.pr.files, n, excluded)
}

func (p suggestingApprover) filterApprover(assignees []string) []string {
	v := make([]string, 0, len(assignees))
	for _, i := range assignees {
		if p.pr.isApprover(i) {
			v = append(v, i)
		}
	}
	return v
}

func (p suggestingApprover) suggestApprover(currentApprovers, assignees []string, log *logrus.Entry) []string {
	f := func(suggested []string) []string {
		n := p.cfg.TotalNumberOfApprovers - len(currentApprovers) - len(suggested)
		if n <= 0 {
			return suggested
		}

		v := p.selectApprovers(mergeSlices(currentApprovers, suggested), n)
		v = append(v, suggested...)
		return v
	}

	as := mergeSlices(currentApprovers, assignees)

	if p.pr.areAllFilesApproved(as, p.cfg.NumberOfApprovers) {
		if len(assignees) > 0 {
			v := p.filterApprover(assignees)
			return f(difference(v, currentApprovers))
		}
		return f([]string{})
	}

	v := p.suggestApproverByNumber(currentApprovers, assignees, log)
	return f(v)
}

func (p suggestingApprover) suggestApproverByNumber(currentApprovers, assignees []string, log *logrus.Entry) []string {
	ah := approverHelper{
		currentApprovers:  currentApprovers,
		assignees:         assignees,
		filenames:         p.pr.files,
		prNumber:          p.pr.info.getNumber(),
		numberOfApprovers: p.cfg.NumberOfApprovers,
		repoOwner:         p.owner,
		prAuthor:          p.pr.prAuthor(),
		allowSelfApprove:  p.cfg.AllowSelfApprove,
		log:               log,
	}
	return ah.suggestApprovers()
}

func mergeSlices(s []string, s1 []string) []string {
	r := make([]string, 0, len(s)+len(s1))
	r = append(r, s...)
	r = append(r, s1...)
	return r
}

func difference(s, s1 []string) []string {
	if len(s) == 0 || len(s1) == 0 {
		return s
	}
	return sets.NewString(s...).Difference(
		sets.NewString(s1...),
	).List()
}

type fakeReviewersClient struct {
	s *suggestingApprover
}

func (f fakeReviewersClient) Reviewers(path string) sets.String {
	return f.s.pr.approversOfFile(path)
}

func (f fakeReviewersClient) LeafReviewers(path string) sets.String {
	return f.s.owner.LeafApprovers(path)
}
