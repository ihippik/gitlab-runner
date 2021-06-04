package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/ihippik/gitlab-runner/config"
)

// GitlabAPI represent API for interacting with Gitlab.
type GitlabAPI struct {
	basePath string
	client   *http.Client
}

// NewGitlabAPI create new  Gitlab API instance.
func NewGitlabAPI(client *http.Client, base string) *GitlabAPI {
	return &GitlabAPI{client: client, basePath: base}
}

// register register new gitlab-runner.
func (g GitlabAPI) register(ctx context.Context, token string, cfg *config.RunnerCfg) (string, error) {
	var regResponse struct {
		ID    int    `json:"id"`
		Token string `json:"token"`
	}

	form := url.Values{}
	form.Add("token", token)
	form.Add("description", cfg.Name)
	form.Add("tag_list", strings.Join(cfg.Tags, ", "))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.basePath+"/runners", strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("new request: %w", err)
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}

	if resp.StatusCode > http.StatusAccepted {
		return "", fmt.Errorf("bad status: %s ", resp.Status)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}

	if err = resp.Body.Close(); err != nil {
		return "", fmt.Errorf("close body: %w", err)
	}

	if err := json.Unmarshal(data, &regResponse); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	return regResponse.Token, nil
}

// jobRequest fetch jobs from Gitlab server.
func (g GitlabAPI) jobRequest(ctx context.Context, jReq *jobRequest) (*jobResponse, error) {
	var jobRequest jobResponse

	reqData, err := json.Marshal(jReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.basePath+"/jobs/request", bytes.NewReader(reqData))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	req.Header.Set("content-type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusCreated:
	case http.StatusNoContent:
		return nil, nil
	case http.StatusForbidden:
		return nil, errors.New("forbidden")
	default:
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if err = resp.Body.Close(); err != nil {
		return nil, fmt.Errorf("close body: %w", err)
	}

	if err := json.Unmarshal(data, &jobRequest); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &jobRequest, nil
}

func (g GitlabAPI) jobTrace(ctx context.Context, startOffset, jobID int, jobToken string, content []byte) (int, error) {
	traceURL := fmt.Sprintf("%s/jobs/%d/trace", g.basePath, jobID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, traceURL, bytes.NewReader(content))
	if err != nil {
		return 0, fmt.Errorf("new request: %w", err)
	}

	endOffset := startOffset + len(content)
	contentRange := fmt.Sprintf("%d-%d", startOffset, endOffset-1)

	req.Header.Set("JOB-TOKEN", jobToken)
	req.Header.Set("Content-Range", contentRange)

	resp, err := g.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("do request: %w", err)
	}

	if err = resp.Body.Close(); err != nil {
		return 0, fmt.Errorf("close body: %w", err)
	}

	if resp.StatusCode != http.StatusAccepted {
		return 0, fmt.Errorf("invalid status: %s", resp.Status)
	}

	return endOffset, nil
}

func (g GitlabAPI) updateJob(ctx context.Context, jobID int, request *updateJobRequest) error {
	data, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	updateURL := fmt.Sprintf("%s/jobs/%d", g.basePath, jobID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, updateURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}

	req.Header.Set("content-type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}

	if err = resp.Body.Close(); err != nil {
		return fmt.Errorf("close body: %w", err)
	}

	if resp.StatusCode > http.StatusAccepted {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	return nil
}

func (g GitlabAPI) uploadArtifacts(ctx context.Context, jobID int, token, path string, options artifactsOptions) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("file open: %w", err)
	}

	fi, err := file.Stat()
	if err != nil {
		return fmt.Errorf("file stat: %w", err)
	}
	defer file.Close()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", fi.Name())
	if err != nil {
		return fmt.Errorf("create form file: %w", err)
	}

	if _, err = io.Copy(part, file); err != nil {
		return fmt.Errorf("copy error: %w", err)
	}

	if err = writer.Close(); err != nil {
		return fmt.Errorf("close writer: %w", err)
	}

	q := url.Values{}

	if options.ExpireIn != "" {
		q.Set("expire_in", options.ExpireIn)
	}

	if options.Format != "" {
		q.Set("artifact_format", string(options.Format))
	}

	if options.Type != "" {
		q.Set("artifact_type", options.Type)
	}

	uploadURL := fmt.Sprintf("%s/jobs/%d/artifacts?%s", g.basePath, jobID, q.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, body)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}

	headers := make(http.Header)
	headers.Set("JOB-TOKEN", token)
	headers.Set("Content-Type", writer.FormDataContentType())
	req.Header = headers

	resp, err := g.client.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode > http.StatusNoContent {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	return nil
}
