// Package runner is a main business layer of service.
package runner

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/ihippik/gitlab-runner/config"
)

// Executor implementation of workers to perform jobs.
type Executor interface {
	Execute(command string) (string, error)
}

// gitlabAPI presents an interface for working with tasks through API Gitlab.
type gitlabAPI interface {
	register(token string, cfg *config.RunnerCfg) (string, error)
	jobRequest(*jobRequest) (*jobResponse, error)
	updateJob(int, *updateJobRequest) error
	jobTrace(startOffset, jobID int, jobToken string, content []byte) (int, error)
}

// Service represent main service struct.
type Service struct {
	logger *logrus.Entry
	config *config.Config

	gitlab   gitlabAPI
	executor Executor

	traceOffset int
}

// NewService create new Service instance.
func NewService(logger *logrus.Entry, config *config.Config, gitlab gitlabAPI, executor Executor) *Service {
	return &Service{logger: logger, config: config, gitlab: gitlab, executor: executor}
}

// Registration register new gitlab-runner in Gitlab.
func (s *Service) Registration(token string) (string, error) {
	runnerToken, err := s.gitlab.register(token, s.config.Runner)
	if err != nil {
		return "", fmt.Errorf("register gitlab-runner: %w", err)
	}

	return runnerToken, nil
}

// Start run main Gitlab-gitlab-runner process.
func (s *Service) Start() error {
	if len(s.config.Runner.Token) == 0 {
		logrus.Warnln("first register and insert the token into the config")
		return nil
	}

	s.logger.WithField("name", s.config.Runner.Name).Infoln("gitlab-runner was started")

	// if job received status changed from pending to running
	job, err := s.gitlab.jobRequest(&jobRequest{
		Token: s.config.Runner.Token,
	})
	if err != nil {
		return fmt.Errorf("job request: %w", err)
	}

	if job == nil {
		s.logger.Debugln("no job")
		return nil
	}

	s.logger.WithFields(logrus.Fields{
		"id":          job.ID,
		"token":       job.Token,
		"steps_count": len(job.Steps),
	}).Infoln("get job")

	helloStr := fmt.Sprintf("Runner %s%s%s greets you!\n", ansiBoldBlue, s.config.Runner.Name, ansiReset)
	s.trace(helloStr, job)

	s.trace("I'm getting started.\n", job)

	if err := s.process(job); err != nil {
		if err := s.jobFailed(job); err != nil {
			return fmt.Errorf("job failed: %w", err)
		}

		return fmt.Errorf("job process: %w", err)
	}

	if err := s.jobFinished(job); err != nil {
		return fmt.Errorf("job finished: %w", err)
	}

	return nil
}

// trace add job trace.
func (s *Service) trace(message string, job *jobResponse) {
	var err error

	s.traceOffset, err = s.gitlab.jobTrace(s.traceOffset, job.ID, job.Token, []byte(message))
	if err != nil {
		s.logger.WithError(err).Errorln("job trace error")
	}
}

// jobFinished set job success state.
func (s *Service) jobFinished(job *jobResponse) error {
	succeeded := fmt.Sprintf("%sJob succeeded!%s", ansiBoldGreen, ansiReset)
	s.trace(succeeded, job)

	if err := s.gitlab.updateJob(
		job.ID,
		&updateJobRequest{
			Token:    job.Token,
			State:    "success",
			ExitCode: 0,
		},
	); err != nil {
		return err
	}

	s.logger.WithField("job_id", job.ID).Infoln("job was finished")

	return nil
}

// jobFailed set job failed state.
func (s *Service) jobFailed(job *jobResponse) error {
	if err := s.gitlab.updateJob(
		job.ID,
		&updateJobRequest{
			Token:         job.Token,
			State:         "failed",
			FailureReason: "script_failure",
			ExitCode:      1,
		},
	); err != nil {
		return err
	}

	s.logger.WithField("job_id", job.ID).Warnln("job failed")

	return nil
}

// process processes all steps of the job.
func (s *Service) process(job *jobResponse) error {
	for _, step := range job.Steps {
		for _, script := range step.Script {
			output, err := s.executor.Execute(script)
			if err != nil {
				return fmt.Errorf("%s: %w", step.Name, err)
			}

			traceStep := fmt.Sprintf("%s%s%s: %s\n", ansiBoldYellow, step.Name, ansiReset, output)
			s.trace(traceStep, job)
		}

		s.logger.WithFields(
			logrus.Fields{"step_name": step.Name, "scripts_count": len(step.Script)},
		).Infoln("step was processed")
	}

	return nil
}
