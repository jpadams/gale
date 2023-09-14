package ghx

import (
	"fmt"
	"os"

	"github.com/aweris/gale/internal/core"
	"github.com/aweris/gale/internal/gctx"
)

// Executor is the interface that defines contract for objects capable of performing an execution task.
type Executor interface {
	// Execute performs the execution of a specific task with the given context.
	Execute(ctx *gctx.Context) error
}

type EnvironmentFiles struct {
	Env         core.EnvironmentFile // Env is the environment file that holds the environment variables
	Path        core.EnvironmentFile // Path is the environment file that holds the path variables
	Outputs     core.EnvironmentFile // Outputs is the environment file that holds the outputs
	StepSummary core.EnvironmentFile // StepSummary is the environment file that holds the step summary
}

func processEnvironmentFiles(ctx *gctx.Context, ef *EnvironmentFiles, ec *gctx.Context) error {
	if ef == nil {
		return nil
	}

	env, err := ef.Env.ReadData(ctx.Context)
	if err != nil {
		return err
	}

	for k, v := range env {
		if err := os.Setenv(k, v); err != nil {
			return err
		}
	}

	paths, err := ef.Path.ReadData(ctx.Context)
	if err != nil {
		return err
	}

	path := os.Getenv("PATH")

	for p := range paths {
		path = fmt.Sprintf("%s:%s", path, p)
	}

	if err := os.Setenv("PATH", path); err != nil {
		return err
	}

	outputs, err := ef.Outputs.ReadData(ctx.Context)
	if err != nil {
		return err
	}

	for k, v := range outputs {
		ec.SetStepOutput(k, v)
	}

	stepSummary, err := ef.StepSummary.RawData(ctx.Context)
	if err != nil {
		return err
	}

	ec.SetStepSummary(stepSummary)

	return nil
}
