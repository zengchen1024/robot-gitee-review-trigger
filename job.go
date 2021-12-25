package main

import (
	"fmt"

	ciparser "github.com/opensourceways/robot-gitee-review-trigger/ci-parser"
)

type jobConfig struct {
	// CITable is the table for CI comment of PR.
	CITable ciparser.CITable `json:"ci_table" required:"true"`

	// JobSuccessStatus is the status desc when a single job is successful
	JobSuccessStatus []string `json:"job_success_status" required:"true"`
}

func (c *jobConfig) validate() error {
	if c == nil {
		return nil
	}

	if err := c.CITable.Validate(); err != nil {
		return err
	}

	if len(c.JobSuccessStatus) == 0 {
		return fmt.Errorf("missing job_success_status")
	}

	return nil
}

func (c jobConfig) newCIParser() ciparser.CIParserImpl {
	return ciparser.CIParserImpl{
		CITable: c.CITable,
		JobStatus: []ciparser.JobStatusDesc{
			{
				Desc:   c.JobSuccessStatus,
				Status: "success",
			},
		},
	}
}

func (c jobConfig) isCISuccess(comment string, jobNumber int) (bool, error) {
	if !c.CITable.IsCIComment(comment) {
		return false, nil
	}

	status, err := ciparser.ParseCIComment(c.newCIParser(), comment)
	if err != nil {
		return false, err
	}

	return len(status) == jobNumber, nil
}
