package repoowners

import (
	"github.com/opensourceways/repo-owners-cache/grpc/client"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/sets"
)

// RepoOwner is an interface to work with repoowners
type RepoOwner interface {
	FindApproverOwnersForFile(path string) string
	FindReviewersOwnersForFile(path string) string
	LeafApprovers(path string) sets.String
	LeafReviewers(path string) sets.String
	Approvers(path string) sets.String
	Reviewers(path string) sets.String
	AllReviewers() sets.String
	TopLevelApprovers() sets.String
}

type owners struct {
	c   *client.Client
	log *logrus.Entry

	org    string
	repo   string
	branch string
}

func NewRepoOwners(org, repo, branch string, c *client.Client, log *logrus.Entry) RepoOwner {
	return &owners{
		org:    org,
		repo:   repo,
		branch: branch,
		c:      c,
		log:    log,
	}
}

func (o *owners) FindApproverOwnersForFile(path string) string {
	return ""
}
func (o *owners) FindReviewersOwnersForFile(path string) string {
	return ""
}
func (o *owners) LeafApprovers(path string) sets.String {
	return sets.NewString()
}
func (o *owners) LeafReviewers(path string) sets.String {
	return sets.NewString()
}
func (o *owners) Approvers(path string) sets.String {
	return sets.NewString()
}
func (o *owners) Reviewers(path string) sets.String {
	return sets.NewString()
}
func (o *owners) AllReviewers() sets.String {
	return sets.NewString()
}
func (o *owners) TopLevelApprovers() sets.String {
	return sets.NewString()
}

type repoMembers struct {
	members sets.String
}

func RepoMemberAsOwners(d []string) RepoOwner {
	return &repoMembers{
		members: sets.NewString(d...),
	}
}

func (o *repoMembers) FindApproverOwnersForFile(path string) string {
	return ""
}
func (o *repoMembers) FindReviewersOwnersForFile(path string) string {
	return ""
}
func (o *repoMembers) LeafApprovers(path string) sets.String {
	return sets.NewString()
}
func (o *repoMembers) LeafReviewers(path string) sets.String {
	return sets.NewString()
}
func (o *repoMembers) Approvers(path string) sets.String {
	return sets.NewString()
}
func (o *repoMembers) Reviewers(path string) sets.String {
	return sets.NewString()
}
func (o *repoMembers) AllReviewers() sets.String {
	return sets.NewString()
}
func (o *repoMembers) TopLevelApprovers() sets.String {
	return sets.NewString()
}
