package runner

type (
	step struct {
		Name         string   `json:"name"`
		Script       []string `json:"script"`
		Timeout      int      `json:"timeout"`
		When         string   `json:"when"`
		AllowFailure bool     `json:"allow_failure"`
	}

	jobResponse struct {
		ID            int    `json:"id"`
		Token         string `json:"token"`
		AllowGitFetch bool   `json:"allow_git_fetch"`
		Steps         []step `json:"steps"`
	}

	jobRequest struct {
		Info       versionInfo `json:"info,omitempty"`
		Token      string      `json:"token,omitempty"`
		LastUpdate string      `json:"last_update,omitempty"`
	}

	versionInfo struct {
		Name         string       `json:"name,omitempty"`
		Version      string       `json:"version,omitempty"`
		Revision     string       `json:"revision,omitempty"`
		Platform     string       `json:"platform,omitempty"`
		Architecture string       `json:"architecture,omitempty"`
		Executor     string       `json:"executor,omitempty"`
		Shell        string       `json:"shell,omitempty"`
		Features     featuresInfo `json:"features"`
	}

	featuresInfo struct {
		Variables               bool `json:"variables"`
		Image                   bool `json:"image"`
		Services                bool `json:"services"`
		Artifacts               bool `json:"artifacts"`
		Cache                   bool `json:"cache"`
		Shared                  bool `json:"shared"`
		UploadMultipleArtifacts bool `json:"upload_multiple_artifacts"`
		UploadRawArtifacts      bool `json:"upload_raw_artifacts"`
		Session                 bool `json:"session"`
		Terminal                bool `json:"terminal"`
		Refspecs                bool `json:"refspecs"`
		Masking                 bool `json:"masking"`
		Proxy                   bool `json:"proxy"`
		RawVariables            bool `json:"raw_variables"`
		ArtifactsExclude        bool `json:"artifacts_exclude"`
		MultiBuildSteps         bool `json:"multi_build_steps"`
		TraceReset              bool `json:"trace_reset"`
		TraceChecksum           bool `json:"trace_checksum"`
		TraceSize               bool `json:"trace_size"`
		VaultSecrets            bool `json:"vault_secrets"`
		Cancelable              bool `json:"cancelable"`
		ReturnExitCode          bool `json:"return_exit_code"`
	}
)

type (
	updateJobRequest struct {
		Info          versionInfo    `json:"info,omitempty"`
		Token         string         `json:"token,omitempty"`
		State         string         `json:"state,omitempty"`
		FailureReason string         `json:"failure_reason,omitempty"`
		Output        jobTraceOutput `json:"output,omitempty"`
		ExitCode      int            `json:"exit_code,omitempty"`
	}

	jobTraceOutput struct {
		Checksum string `json:"checksum,omitempty"`
		Bytesize int    `json:"bytesize,omitempty"`
	}
)
