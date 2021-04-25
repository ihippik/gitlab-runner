package main

import (
	"fmt"
	"io/ioutil"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/ihippik/gitlab-runner/config"
	"github.com/ihippik/gitlab-runner/executor"
	"github.com/ihippik/gitlab-runner/runner"
)

// initLogger init logrus logger with specified fields and log level.
func initLogger(cfg *config.LoggerCfg, version, executor string) *logrus.Entry {
	logger := logrus.New()

	switch cfg.Level {
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "warn":
		logger.SetLevel(logrus.WarnLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	return logger.WithFields(logrus.Fields{"version": version, "executor": executor})
}

// initConfig parse config from yaml file.
func initConfig(path string) (*config.Config, error) {
	var cfg config.Config

	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	if err := yaml.Unmarshal(yamlFile, &cfg); err != nil {
		return nil, fmt.Errorf("parse file: %w", err)
	}

	return &cfg, nil
}

func executorFactory(logger *logrus.Entry, kind string) runner.Executor {
	switch kind {
	case "shell":
		return executor.NewShellExecutor()
	default:
		logger.WithField("executor_kind", kind).Fatalln("not support yet")
		return nil
	}
}
