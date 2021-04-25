// Package config provides configuration service structure and utilities.
package config

type (
	// Config represent service config.
	Config struct {
		Runner *RunnerCfg `json:"gitlab-runner"`
		Logger *LoggerCfg `json:"logger"`
	}

	// RunnerCfg gitlab-runner config section.
	RunnerCfg struct {
		Name     string   `json:"name,omitempty"`
		URL      string   `json:"url,omitempty"`
		Token    string   `json:"token,omitempty"`
		Executor string   `json:"executor,omitempty"`
		Tags     []string `json:"tags"`
	}

	// LoggerCfg logger config section.
	LoggerCfg struct {
		Level string `json:"level,omitempty"`
	}
)
