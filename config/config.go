// Package config provides configuration service structure and utilities.
package config

import "time"

type (
	// Config represent service config.
	Config struct {
		Runner *RunnerCfg
		Logger *LoggerCfg
	}

	// RunnerCfg gitlab-runner config section.
	RunnerCfg struct {
		Name     string
		URL      string
		Token    string
		Executor string
		Tags     []string
		Interval time.Duration
	}

	// LoggerCfg logger config section.
	LoggerCfg struct {
		Level string
	}
)
