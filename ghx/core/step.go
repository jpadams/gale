package core

import "strings"

// Step represents a single task in a job context at GitHub Actions workflow
//
// See: https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idsteps
type Step struct {
	ID               string            `yaml:"id,omitempty"`                // ID is the unique identifier of the step.
	If               string            `yaml:"if,omitempty"`                // If is the conditional expression to run the step.
	Name             string            `yaml:"name,omitempty"`              // Name is the name of the step.
	Uses             string            `yaml:"uses,omitempty"`              // Uses is the action to run for the step.
	Environment      map[string]string `yaml:"env,omitempty"`               // Environment maps environment variable names to their values.
	With             map[string]string `yaml:"with,omitempty"`              // With maps input names to their values for the step.
	Run              string            `yaml:"run,omitempty"`               // Run is the command to run for the step.
	Shell            string            `yaml:"shell,omitempty"`             // Shell is the shell to use for the step.
	WorkingDirectory string            `yaml:"working-directory,omitempty"` // WorkingDirectory is the working directory for the step.
	ContinueOnError  bool              `yaml:"continue-on-error,omitempty"` // ContinueOnError is a flag to continue on error.
	TimeoutMinutes   int               `yaml:"timeout-minutes,omitempty"`   // TimeoutMinutes is the maximum number of minutes to run the step.
}

// Type returns the type of the step according to its properties
func (s *Step) Type() StepType {
	var st StepType

	// determine the type of the step based on its properties
	switch {
	case s.Uses != "" && strings.HasPrefix(s.Uses, "docker://"):
		st = StepTypeDocker
	case s.Uses != "":
		st = StepTypeAction
	case s.Run != "":
		st = StepTypeRun
	default:
		st = StepTypeUnknown
	}

	return st
}

// StepRun represents a single job run in a GitHub Actions workflow run
type StepRun struct {
	Step        Step              `json:"step"`        // Step is the step to run
	Stage       StepStage         `json:"stage"`       // Stage is the stage of the step during the execution of the job. Possible values are: setup, pre, main, post, complete.
	Conclusion  Conclusion        `json:"conclusion"`  // Conclusion is the result of a completed job after continue-on-error is applied
	Outcome     Conclusion        `json:"outcome"`     // Outcome is  the result of a completed job before continue-on-error is applied
	Outputs     map[string]string `json:"outputs"`     // Outputs is the outputs generated by the job
	State       map[string]string `json:"state"`       // State is a map of step state variables.
	Summary     string            `json:"summary"`     // Summary is the summary of the step.
	Environment map[string]string `json:"environment"` // Environment is the extra environment variables set by the step.
	Path        []string          `json:"path"`        // Path is extra PATH items set by the step.
}
