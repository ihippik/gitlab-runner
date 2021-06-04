package runner

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type ExecutorMock struct {
	mock.Mock
}

func (e *ExecutorMock) HomeDirectory(dir string) {
	panic("implement me")
}

func (e *ExecutorMock) Execute(ctx context.Context, command string) (string, error) {
	args := e.Called(ctx, command)
	return args.String(0), args.Error(1)
}
