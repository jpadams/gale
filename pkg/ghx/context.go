package ghx

import (
	"fmt"
	"math"
	"os"
	"path/filepath"

	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/internal/fs"

	"github.com/aweris/gale/internal/core"
	"github.com/aweris/gale/internal/expression"
)

var _ expression.VariableProvider = new(ExprContext)

type ExprContext struct {
	Github  core.GithubContext
	Runner  core.RunnerContext
	Job     core.JobContext
	Steps   map[string]core.StepContext
	Secrets core.SecretsContext
	Inputs  map[string]string

	// TODO: add other contexts when needed.
	//  - env context
	//  - vars context
	//  - strategy context
	//  - matrix context
	//  - needs context
	//  - jobs context
}

func NewExprContext() (*ExprContext, error) {
	path := filepath.Join(config.GhxHome(), "secrets", "secrets.json")

	err := fs.EnsureFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure secrets file exist: %w", err)
	}

	var secrets core.SecretsContext

	err = fs.ReadJSONFile(path, &secrets)
	if err != nil {
		return nil, fmt.Errorf("failed to read secrets file: %w", err)
	}

	gc, err := LoadGithubContextFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to create github context: %w", err)
	}

	return &ExprContext{
		Github: *gc,
		Runner: core.RunnerContext{
			Name:      os.Getenv("RUNNER_NAME"),
			OS:        os.Getenv("RUNNER_OS"),
			Arch:      os.Getenv("RUNNER_ARCH"),
			Temp:      os.Getenv("RUNNER_TEMP"),
			ToolCache: os.Getenv("RUNNER_TOOL_CACHE"),
			Debug:     os.Getenv("RUNNER_DEBUG"),
		},
		Job: core.JobContext{
			Status: core.ConclusionSuccess, // start with success status
		},
		Steps:   make(map[string]core.StepContext),
		Secrets: secrets,
		Inputs:  make(map[string]string),
	}, nil
}

func LoadGithubContextFromEnv() (*core.GithubContext, error) {
	// event data
	var event map[string]interface{}

	err := fs.ReadJSONFile(os.Getenv("GITHUB_EVENT_PATH"), &event)
	if err != nil {
		return nil, fmt.Errorf("failed to read event file: %w", err)
	}

	gc := &core.GithubContext{
		Repository:        os.Getenv("GITHUB_REPOSITORY"),
		RepositoryID:      os.Getenv("GITHUB_REPOSITORY_ID"),
		RepositoryOwner:   os.Getenv("GITHUB_REPOSITORY_OWNER"),
		RepositoryOwnerID: os.Getenv("GITHUB_REPOSITORY_OWNER_ID"),
		RepositoryURL:     os.Getenv("GITHUB_REPOSITORY_URL"),
		Workspace:         os.Getenv("GITHUB_WORKSPACE"),
		APIURL:            os.Getenv("GITHUB_API_URL"),
		GraphqlURL:        os.Getenv("GITHUB_GRAPHQL_URL"),
		ServerURL:         os.Getenv("GITHUB_SERVER_URL"),
		Ref:               os.Getenv("GITHUB_REF"),
		RefName:           os.Getenv("GITHUB_REF_NAME"),
		RefType:           os.Getenv("GITHUB_REF_TYPE"),
		RefProtected:      os.Getenv("GITHUB_REF_PROTECTED") == "true",
		HeadRef:           os.Getenv("GITHUB_HEAD_REF"),
		BaseRef:           os.Getenv("GITHUB_BASE_REF"),
		SHA:               os.Getenv("GITHUB_SHA"),
		EventName:         os.Getenv("GITHUB_EVENT_NAME"),
		EventPath:         os.Getenv("GITHUB_EVENT_PATH"),
		Token:             os.Getenv("GITHUB_TOKEN"),
		Event:             event,
	}

	return gc, nil
}

func (c *ExprContext) GetVariable(name string) (interface{}, error) {
	switch name {
	case "github":
		return c.Github, nil
	case "runner":
		return c.Runner, nil
	case "env":
		return map[string]string{}, nil
	case "vars":
		return map[string]string{}, nil
	case "job":
		return c.Job, nil
	case "steps":
		return c.Steps, nil
	case "secrets":
		return c.Secrets, nil
	case "strategy":
		return map[string]string{}, nil
	case "matrix":
		return map[string]string{}, nil
	case "needs":
		return map[string]string{}, nil
	case "inputs":
		return c.Inputs, nil
	case "infinity":
		return math.Inf(1), nil
	case "nan":
		return math.NaN(), nil
	default:
		return nil, fmt.Errorf("unknown variable: %s", name)
	}
}

// WithGithubEnv sets `github.env` from the given environment file. This is path of the temporary file that holds the
// environment variables
func (c *ExprContext) WithGithubEnv(path string) *ExprContext {
	c.Github.Env = path

	return c
}

// WithoutGithubEnv removes `github.env` from the context.
func (c *ExprContext) WithoutGithubEnv() *ExprContext {
	c.Github.Env = ""

	return c
}

// WithGithubPath sets `github.path` from the given environment file. This is path of the temporary file that holds the
func (c *ExprContext) WithGithubPath(path string) *ExprContext {
	c.Github.Path = path

	return c
}

// WithoutGithubPath removes `github.path` from the context.
func (c *ExprContext) WithoutGithubPath() *ExprContext {
	c.Github.Path = ""
	return c
}

// SetStepOutput sets the output of the given step.
func (c *ExprContext) SetStepOutput(stepID, key, value string) *ExprContext {
	sc, ok := c.Steps[stepID]
	if !ok {
		sc = core.StepContext{}
	}

	if sc.Outputs == nil {
		sc.Outputs = make(map[string]string)
	}

	sc.Outputs[key] = value

	c.Steps[stepID] = sc

	return c
}

// SetStepResult sets the result of the given step.
func (c *ExprContext) SetStepResult(stepID string, outcome, conclusion core.Conclusion) *ExprContext {
	sc, ok := c.Steps[stepID]
	if !ok {
		sc = core.StepContext{}
	}

	sc.Outcome = outcome
	sc.Conclusion = conclusion

	c.Steps[stepID] = sc

	return c
}

// SetStepSummary sets the summary of the given step.
func (c *ExprContext) SetStepSummary(stepID, summary string) *ExprContext {
	sc, ok := c.Steps[stepID]
	if !ok {
		sc = core.StepContext{}
	}

	sc.Summary = summary

	c.Steps[stepID] = sc

	return c
}

// SetStepState sets the state of the given step.
func (c *ExprContext) SetStepState(stepID, key, value string) *ExprContext {
	sc, ok := c.Steps[stepID]
	if !ok {
		sc = core.StepContext{}
	}

	if sc.State == nil {
		sc.State = make(map[string]string)
	}

	sc.State[key] = value

	c.Steps[stepID] = sc

	return c
}

// SetJobStatus sets the status of the job.
func (c *ExprContext) SetJobStatus(status core.Conclusion) *ExprContext {
	c.Job.Status = status

	return c
}