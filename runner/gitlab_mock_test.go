package runner

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/ihippik/gitlab-runner/config"
)

type GitlabAPIMock struct {
	mock.Mock
}

func (g *GitlabAPIMock) uploadArtifacts(
	ctx context.Context,
	id int,
	token, path string,
	options artifactsOptions,
) error {
	panic("implement me")
}

func (g *GitlabAPIMock) register(ctx context.Context, token string, cfg *config.RunnerCfg) (string, error) {
	args := g.Called(ctx, token, cfg)
	return args.String(0), args.Error(1)
}

func (g *GitlabAPIMock) jobRequest(ctx context.Context, req *jobRequest) (*jobResponse, error) {
	args := g.Called(ctx, req)
	return args.Get(0).(*jobResponse), args.Error(1)
}

func (g *GitlabAPIMock) updateJob(ctx context.Context, id int, req *updateJobRequest) error {
	args := g.Called(ctx, id, req)
	return args.Error(0)
}

func (g *GitlabAPIMock) jobTrace(
	ctx context.Context,
	startOffset,
	jobID int,
	jobToken string,
	content []byte,
) (int, error) {
	args := g.Called(ctx, startOffset, jobID, jobToken, content)
	return args.Int(0), args.Error(1)
}
