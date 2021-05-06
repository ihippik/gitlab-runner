package runner

import (
	"context"
	"errors"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ihippik/gitlab-runner/config"
)

func TestService_Registration(t *testing.T) {
	type (
		fields struct {
			config *config.Config
		}
		args struct {
			token string
		}
	)

	gitlab := new(GitlabAPIMock)
	setRegister := func(token string, cfg *config.RunnerCfg, result string, err error) {
		gitlab.On("register", mock.Anything, token, cfg).Return(result, err).Once()
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		setup   func()
		want    string
		wantErr error
	}{
		{
			name: "success",
			fields: fields{
				config: &config.Config{
					Runner: &config.RunnerCfg{
						Name:     "my",
						URL:      "https://gitlab.com",
						Token:    "qwerty",
						Executor: "shell",
						Tags:     []string{"custom"},
						Interval: 10,
					},
				},
			},
			args: args{
				token: "asd",
			},
			setup: func() {
				setRegister(
					"asd",
					&config.RunnerCfg{
						Name:     "my",
						URL:      "https://gitlab.com",
						Token:    "qwerty",
						Executor: "shell",
						Tags:     []string{"custom"},
						Interval: 10,
					},
					"token-res",
					nil,
				)
			},
			want:    "token-res",
			wantErr: nil,
		},
		{
			name: "register error",
			fields: fields{
				config: &config.Config{
					Runner: &config.RunnerCfg{
						Name:     "my",
						URL:      "https://gitlab.com",
						Token:    "qwerty",
						Executor: "shell",
						Tags:     []string{"custom"},
						Interval: 10,
					},
				},
			},
			args: args{
				token: "asd",
			},
			setup: func() {
				setRegister(
					"asd",
					&config.RunnerCfg{
						Name:     "my",
						URL:      "https://gitlab.com",
						Token:    "qwerty",
						Executor: "shell",
						Tags:     []string{"custom"},
						Interval: 10,
					},
					"token-res",
					errors.New("some err"),
				)
			},
			want:    "",
			wantErr: errors.New("register gitlab-runner: some err"),
		},
	}

	logger, _ := test.NewNullLogger()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer gitlab.AssertExpectations(t)
			tt.setup()

			s := &Service{
				logger: logrus.NewEntry(logger),
				config: tt.fields.config,
				gitlab: gitlab,
			}

			got, err := s.Registration(context.Background(), tt.args.token)
			if err != nil && assert.Error(t, tt.wantErr) {
				assert.EqualError(t, err, tt.wantErr.Error())
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_processJob(t *testing.T) {
	type fields struct {
		config *config.Config
	}

	logger, _ := test.NewNullLogger()
	gitlab := new(GitlabAPIMock)
	executor := new(ExecutorMock)

	setJobRequest := func(req *jobRequest, resp *jobResponse, err error) {
		gitlab.On("jobRequest", mock.Anything, req).Return(resp, err).Once()
	}

	setUpdateJob := func(id int, req *updateJobRequest, err error) {
		gitlab.On("updateJob", mock.Anything, id, req).Return(err).Once()
	}

	setJobTrace := func(startOffset, jobID int, jobToken string, content []byte, result int, err error) {
		gitlab.On(
			"jobTrace",
			mock.Anything,
			startOffset,
			jobID,
			jobToken,
			content,
		).Return(result, err).Once()
	}

	setExecutor := func(command, output string, err error) {
		executor.On("Execute", mock.Anything, command).Return(output, err).Once()
	}

	tests := []struct {
		name            string
		wantTraceOffset int
		wantError       error
		fields          fields
		setup           func()
	}{
		{
			name: "success",
			fields: fields{
				config: &config.Config{
					Runner: &config.RunnerCfg{
						Name:     "my-runner",
						URL:      "",
						Token:    "my-token",
						Executor: "",
						Tags:     nil,
						Interval: 0,
					},
				},
			},
			wantTraceOffset: 40,
			setup: func() {
				setJobRequest(
					&jobRequest{Token: "my-token"},
					&jobResponse{
						ID:    2,
						Token: "job-token",
						Steps: []step{
							{
								Name:         "step-name",
								Script:       []string{"command"},
								Timeout:      0,
								When:         "",
								AllowFailure: false,
							},
						},
					},
					nil,
				)

				setJobTrace(
					0,
					2,
					"job-token",
					[]byte{0x52, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x20, 0x1b, 0x5b, 0x33, 0x34, 0x3b, 0x31, 0x6d, 0x6d, 0x79, 0x2d, 0x72, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x1b, 0x5b, 0x30, 0x3b, 0x6d, 0x20, 0x67, 0x72, 0x65, 0x65, 0x74, 0x73, 0x20, 0x79, 0x6f, 0x75, 0x21, 0xa},
					10,
					nil,
				)

				setJobTrace(
					10,
					2,
					"job-token",
					[]byte{0x49, 0x27, 0x6d, 0x20, 0x67, 0x65, 0x74, 0x74, 0x69, 0x6e, 0x67, 0x20, 0x73, 0x74, 0x61, 0x72, 0x74, 0x65, 0x64, 0x2e, 0xa},
					20,
					nil,
				)

				setExecutor("command", "hello!", nil)

				setJobTrace(
					20,
					2,
					"job-token",
					[]byte{0x1b, 0x5b, 0x33, 0x33, 0x3b, 0x31, 0x6d, 0x73, 0x74, 0x65, 0x70, 0x2d, 0x6e, 0x61, 0x6d, 0x65, 0x1b, 0x5b, 0x30, 0x3b, 0x6d, 0x3a, 0x20, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x21, 0xa},
					30,
					nil,
				)

				setJobTrace(
					30,
					2,
					"job-token",
					[]byte{0x1b, 0x5b, 0x33, 0x32, 0x3b, 0x31, 0x6d, 0x4a, 0x6f, 0x62, 0x20, 0x73, 0x75, 0x63, 0x63, 0x65, 0x65, 0x64, 0x65, 0x64, 0x21, 0x1b, 0x5b, 0x30, 0x3b, 0x6d},
					40,
					nil,
				)

				setUpdateJob(
					2,
					&updateJobRequest{
						Token:         "job-token",
						State:         "success",
						FailureReason: "",
						Output:        jobTraceOutput{},
						ExitCode:      0,
					},
					nil,
				)
			},
		},
		{
			name: "request jobs error",
			fields: fields{
				config: &config.Config{
					Runner: &config.RunnerCfg{
						Name:     "my-runner",
						URL:      "",
						Token:    "my-token",
						Executor: "",
						Tags:     nil,
						Interval: 0,
					},
				},
			},
			wantError:       errors.New("job request: some error"),
			wantTraceOffset: 0,
			setup: func() {
				setJobRequest(
					&jobRequest{Token: "my-token"},
					&jobResponse{
						ID:    2,
						Token: "job-token",
						Steps: []step{
							{
								Name:         "step-name",
								Script:       []string{"command"},
								Timeout:      0,
								When:         "",
								AllowFailure: false,
							},
						},
					},
					errors.New("some error"),
				)
			},
		},
		{
			name: "no job",
			fields: fields{
				config: &config.Config{
					Runner: &config.RunnerCfg{
						Name:     "my-runner",
						URL:      "",
						Token:    "my-token",
						Executor: "",
						Tags:     nil,
						Interval: 0,
					},
				},
			},
			wantTraceOffset: 0,
			setup: func() {
				setJobRequest(
					&jobRequest{Token: "my-token"},
					nil,
					nil,
				)
			},
		},
		{
			name: "executor error",
			fields: fields{
				config: &config.Config{
					Runner: &config.RunnerCfg{
						Name:     "my-runner",
						URL:      "",
						Token:    "my-token",
						Executor: "",
						Tags:     nil,
						Interval: 0,
					},
				},
			},
			wantError:       errors.New("job process: step-name: some err"),
			wantTraceOffset: 20,
			setup: func() {
				setJobRequest(
					&jobRequest{Token: "my-token"},
					&jobResponse{
						ID:    2,
						Token: "job-token",
						Steps: []step{
							{
								Name:         "step-name",
								Script:       []string{"command"},
								Timeout:      0,
								When:         "",
								AllowFailure: false,
							},
						},
					},
					nil,
				)

				setJobTrace(
					0,
					2,
					"job-token",
					[]byte{0x52, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x20, 0x1b, 0x5b, 0x33, 0x34, 0x3b, 0x31, 0x6d, 0x6d, 0x79, 0x2d, 0x72, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x1b, 0x5b, 0x30, 0x3b, 0x6d, 0x20, 0x67, 0x72, 0x65, 0x65, 0x74, 0x73, 0x20, 0x79, 0x6f, 0x75, 0x21, 0xa},
					10,
					nil,
				)

				setJobTrace(
					10,
					2,
					"job-token",
					[]byte{0x49, 0x27, 0x6d, 0x20, 0x67, 0x65, 0x74, 0x74, 0x69, 0x6e, 0x67, 0x20, 0x73, 0x74, 0x61, 0x72, 0x74, 0x65, 0x64, 0x2e, 0xa},
					20,
					nil,
				)

				setExecutor("command", "hello!", errors.New("some err"))

				setUpdateJob(
					2,
					&updateJobRequest{
						Token:         "job-token",
						State:         "failed",
						FailureReason: "script_failure",
						Output:        jobTraceOutput{},
						ExitCode:      1,
					},
					nil,
				)
			},
		},
		{
			name: "executor error: job failed",
			fields: fields{
				config: &config.Config{
					Runner: &config.RunnerCfg{
						Name:     "my-runner",
						URL:      "",
						Token:    "my-token",
						Executor: "",
						Tags:     nil,
						Interval: 0,
					},
				},
			},
			wantError:       errors.New("process: job failed: some update job err"),
			wantTraceOffset: 20,
			setup: func() {
				setJobRequest(
					&jobRequest{Token: "my-token"},
					&jobResponse{
						ID:    2,
						Token: "job-token",
						Steps: []step{
							{
								Name:         "step-name",
								Script:       []string{"command"},
								Timeout:      0,
								When:         "",
								AllowFailure: false,
							},
						},
					},
					nil,
				)

				setJobTrace(
					0,
					2,
					"job-token",
					[]byte{0x52, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x20, 0x1b, 0x5b, 0x33, 0x34, 0x3b, 0x31, 0x6d, 0x6d, 0x79, 0x2d, 0x72, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x1b, 0x5b, 0x30, 0x3b, 0x6d, 0x20, 0x67, 0x72, 0x65, 0x65, 0x74, 0x73, 0x20, 0x79, 0x6f, 0x75, 0x21, 0xa},
					10,
					nil,
				)

				setJobTrace(
					10,
					2,
					"job-token",
					[]byte{0x49, 0x27, 0x6d, 0x20, 0x67, 0x65, 0x74, 0x74, 0x69, 0x6e, 0x67, 0x20, 0x73, 0x74, 0x61, 0x72, 0x74, 0x65, 0x64, 0x2e, 0xa},
					20,
					nil,
				)

				setExecutor("command", "hello!", errors.New("some err"))

				setUpdateJob(
					2,
					&updateJobRequest{
						Token:         "job-token",
						State:         "failed",
						FailureReason: "script_failure",
						Output:        jobTraceOutput{},
						ExitCode:      1,
					},
					errors.New("some update job err"),
				)
			},
		},
		{
			name: "success: update job error",
			fields: fields{
				config: &config.Config{
					Runner: &config.RunnerCfg{
						Name:     "my-runner",
						URL:      "",
						Token:    "my-token",
						Executor: "",
						Tags:     nil,
						Interval: 0,
					},
				},
			},
			wantError:       errors.New("job finished: some err"),
			wantTraceOffset: 40,
			setup: func() {
				setJobRequest(
					&jobRequest{Token: "my-token"},
					&jobResponse{
						ID:    2,
						Token: "job-token",
						Steps: []step{
							{
								Name:         "step-name",
								Script:       []string{"command"},
								Timeout:      0,
								When:         "",
								AllowFailure: false,
							},
						},
					},
					nil,
				)

				setJobTrace(
					0,
					2,
					"job-token",
					[]byte{0x52, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x20, 0x1b, 0x5b, 0x33, 0x34, 0x3b, 0x31, 0x6d, 0x6d, 0x79, 0x2d, 0x72, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x1b, 0x5b, 0x30, 0x3b, 0x6d, 0x20, 0x67, 0x72, 0x65, 0x65, 0x74, 0x73, 0x20, 0x79, 0x6f, 0x75, 0x21, 0xa},
					10,
					nil,
				)

				setJobTrace(
					10,
					2,
					"job-token",
					[]byte{0x49, 0x27, 0x6d, 0x20, 0x67, 0x65, 0x74, 0x74, 0x69, 0x6e, 0x67, 0x20, 0x73, 0x74, 0x61, 0x72, 0x74, 0x65, 0x64, 0x2e, 0xa},
					20,
					nil,
				)

				setExecutor("command", "hello!", nil)

				setJobTrace(
					20,
					2,
					"job-token",
					[]byte{0x1b, 0x5b, 0x33, 0x33, 0x3b, 0x31, 0x6d, 0x73, 0x74, 0x65, 0x70, 0x2d, 0x6e, 0x61, 0x6d, 0x65, 0x1b, 0x5b, 0x30, 0x3b, 0x6d, 0x3a, 0x20, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x21, 0xa},
					30,
					nil,
				)

				setJobTrace(
					30,
					2,
					"job-token",
					[]byte{0x1b, 0x5b, 0x33, 0x32, 0x3b, 0x31, 0x6d, 0x4a, 0x6f, 0x62, 0x20, 0x73, 0x75, 0x63, 0x63, 0x65, 0x65, 0x64, 0x65, 0x64, 0x21, 0x1b, 0x5b, 0x30, 0x3b, 0x6d},
					40,
					nil,
				)

				setUpdateJob(
					2,
					&updateJobRequest{
						Token:         "job-token",
						State:         "success",
						FailureReason: "",
						Output:        jobTraceOutput{},
						ExitCode:      0,
					},
					errors.New("some err"),
				)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			defer gitlab.AssertExpectations(t)
			defer executor.AssertExpectations(t)

			s := &Service{
				logger:      logrus.NewEntry(logger),
				config:      tt.fields.config,
				gitlab:      gitlab,
				executor:    executor,
				errChan:     make(chan error, 100),
				traceOffset: 0,
			}
			s.processJob(context.Background())
			select {
			case err := <-s.errChan:
				if assert.NotNil(t, tt.wantError) {
					assert.EqualError(t, err, tt.wantError.Error())
				}
			default:
				assert.NoError(t, tt.wantError)
			}

			assert.Equal(t, tt.wantTraceOffset, s.traceOffset)
		})
	}
}
