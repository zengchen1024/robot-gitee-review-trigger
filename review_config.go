package main

import (
	"fmt"
	"regexp"

	"k8s.io/apimachinery/pkg/util/sets"
)

type ownerConfig struct {
	// BranchWithoutOwners is a list of branches which have no OWNERS file
	// For these branch, collaborators will work as the approvers
	// It can't be set with BranchWithOwners at same time
	BranchWithoutOwners   string `json:"branch_without_owners"`
	reBranchWithoutOwners *regexp.Regexp

	// BranchWithOwners is a list of branches which have OWNERS file
	// It can't be set with BranchWithoutOwners at same time
	BranchWithOwners []string `json:"branch_with_owners"`
}

func (o *ownerConfig) setDefault() {}

func (o *ownerConfig) validate() error {
	if o == nil {
		return nil
	}

	if len(o.BranchWithOwners) > 0 && o.BranchWithoutOwners != "" {
		return fmt.Errorf("both `branch_with_owners` and `branch_without_owners` can not be set at same time")
	}

	if o.BranchWithoutOwners != "" {
		r, err := regexp.Compile(o.BranchWithoutOwners)
		if err != nil {
			return fmt.Errorf("the value of `branch_without_owners` is not a valid regexp, err:%v", err)
		}
		o.reBranchWithoutOwners = r
	}
	return nil
}

// len(o.BranchWithOwners) == 0 && o.BranchWithoutOwners == "" means all branch has OWNERS file.
func (o ownerConfig) IsBranchWithoutOwners(branch string) bool {
	if len(o.BranchWithOwners) > 0 {
		return !sets.NewString(o.BranchWithOwners...).Has(branch)
	}

	return o.reBranchWithoutOwners != nil && o.reBranchWithoutOwners.MatchString(branch)
}

type reviewConfig struct {
	// AllowSelfApprove is the tag which indicate if the author
	// can appove his/her own pull-request.
	AllowSelfApprove bool `json:"allow_self_approve"`

	// NumberOfApprovers is the min number of approvers who commented
	// /approve at same time to approve a single module
	NumberOfApprovers int `json:"number_of_approvers"`

	// TotalNumberOfApprovers is the min number of approvers who commented
	// /approve at same time to add approved label
	TotalNumberOfApprovers int `json:"total_number_of_approvers"`

	// TotalNumberOfReviewers is the min number of reviewers who commented
	// /lgtm at same time to add lgtm label
	TotalNumberOfReviewers int `json:"total_number_of_reviewers"`
}

func (r reviewConfig) validate() error {
	return nil
}

func (r *reviewConfig) setDefault() {
	if r == nil {
		return
	}

	if r.NumberOfApprovers <= 0 {
		r.NumberOfApprovers = 1
	}

	if r.TotalNumberOfApprovers <= 0 {
		r.TotalNumberOfApprovers = 2
	}

	if r.TotalNumberOfReviewers == 0 {
		r.TotalNumberOfReviewers = 1
	}
}
