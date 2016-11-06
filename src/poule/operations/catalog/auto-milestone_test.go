package catalog

import (
	"testing"

	"poule/gh"
	"poule/operations"
	"poule/test"

	"github.com/google/go-github/github"
	"github.com/stretchr/testify/mock"
)

func TestAutoMilestone(t *testing.T) {
	clt, ctx := makeContext()
	operation := &autoMilestoneOperation{
		VersionGetter: func(repository string) (string, error) {
			return "test-version", nil
		},
	}

	// Create test pull request and related issue object.
	pullr := test.NewPullRequestBuilder(test.IssueNumber).
		Merged(true).
		HeadBranch(ctx.Username, ctx.Repository, "head", test.CommitSHA[0]).
		BaseBranch(ctx.Username, ctx.Repository, "master", test.CommitSHA[1]).Value

	// Mock the milestones API.
	milestones := []*github.Milestone{
		&github.Milestone{
			Number: test.MakeInt(1),
			Title:  test.MakeString("old-version"),
			State:  test.MakeString("closed"),
		},
		&github.Milestone{
			Number: test.MakeInt(2),
			Title:  test.MakeString("other-version"),
			State:  test.MakeString("open"),
		},
		&github.Milestone{
			Number: test.MakeInt(3),
			Title:  test.MakeString("test-version"),
			State:  test.MakeString("open"),
		},
	}
	clt.MockIssues.On("ListMilestones", ctx.Username, ctx.Repository, mock.AnythingOfType("*github.MilestoneListOptions")).
		Return(milestones, nil, nil)
	clt.MockIssues.On("Edit", ctx.Username, ctx.Repository, test.IssueNumber, mock.AnythingOfType("*github.IssueRequest")).
		Run(func(args mock.Arguments) {
			arg := args.Get(3).(*github.IssueRequest)
			if arg.Milestone == nil {
				t.Fatalf("Expected milestone to be updated")
			}
			if *arg.Milestone != 3 {
				t.Fatalf("Expected milestone 3 to be set, got %d.", *arg.Milestone)
			}
		}).
		Return(nil, nil, nil)

	// Call into the operation.
	item := gh.MakePullRequestItem(pullr)
	res, userData, err := operation.Filter(ctx, item)
	if err != nil {
		t.Fatalf("Filter returned unexpected error %v", err)
	}
	if res != operations.Accept {
		t.Fatalf("Filter returned unexpected result %v", res)
	}
	if err := operation.Apply(ctx, item, userData); err != nil {
		t.Fatalf("Apply returned unexpected error %v", err)
	}
	test.AssertExpectations(clt, t)
}