package main

import (
	"fmt"

	"github.com/opensourceways/go-gitee/gitee"
	"github.com/sirupsen/logrus"
)

type ciConfig struct {
	// NoCI is the tag which indicates the repo is not set CI.
	NoCI bool `json:"no_ci,omitempty"`

	Job *jobConfig `json:"job,omitempty"`

	// NumberOfTestCases is the number of test cases for PR
	NumberOfTestCases int `json:"number_of_test_cases,omitempty"`

	// LabelForCIPassed is the label name for org/repos indicating
	// the CI test cases have passed
	LabelForCIPassed string `json:"label_for_ci_passed,omitempty"`
}

func (c *ciConfig) setDefault() {}

func (c *ciConfig) validate() error {
	if c == nil {
		return nil
	}

	if c.NoCI {
		return nil
	}

	if c.NumberOfTestCases <= 0 {
		return fmt.Errorf("number_of_test_cases must be begger than 0")
	}

	if c.LabelForCIPassed == "" {
		return fmt.Errorf("missing label_for_ci_passed")
	}

	if c.Job == nil {
		return fmt.Errorf("missing job")
	}

	return c.Job.validate()
}

func canHandleCIEvent(e *gitee.NoteEvent, cfg ciConfig) (bool, error) {
	if cfg.NoCI {
		return false, nil
	}

	return cfg.Job.isCISuccess(e.GetComment().GetBody(), cfg.NumberOfTestCases)
}

func (rt *robot) handleCIStatusComment(e *gitee.NoteEvent, cfg *botConfig, log *logrus.Entry) error {
	if b, err := canHandleCIEvent(e, cfg.CI); !b {
		return err
	}

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

	info, err := rt.getReviewInfo(prInfo)
	if err != nil {
		return err
	}

	stats := &reviewStats{
		pr:        &pr,
		cfg:       cfg.Review,
		reviewers: owner.AllReviewers(),
	}

	rs, r := info.doStats(stats, rt.botName)

	if rs.IsEmpty() {
		return rt.readyToReview(prInfo, cfg, log)
	}

	pa := PostAction{
		c:                rt.client,
		cfg:              cfg,
		owner:            owner,
		log:              log,
		pr:               &pr,
		isStartingReview: true,
	}

	return pa.do(info.reviewGuides(rt.botName), "", rs, r, rt.botName)
}

type prInfoOnNoteEvent struct {
	e *gitee.NoteEvent
}

func (pr prInfoOnNoteEvent) getOrgAndRepo() (string, string) {
	return pr.e.GetOrgRepo()
}

func (pr prInfoOnNoteEvent) getNumber() int32 {
	return pr.e.GetPRNumber()
}

func (pr prInfoOnNoteEvent) getTargetBranch() string {
	return pr.e.GetPRBaseRef()
}

func (pr prInfoOnNoteEvent) hasLabel(l string) bool {
	return pr.e.GetPRLabelSet().Has(l)
}
func (pr prInfoOnNoteEvent) getAuthor() string {
	return pr.e.GetPRAuthor()
}

func (pr prInfoOnNoteEvent) getHeadSHA() string {
	return pr.e.GetPRHeadSha()
}
