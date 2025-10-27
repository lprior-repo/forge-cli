package ci

import (
	"fmt"
	"os"
	"strconv"
)

// Environment represents a CI/CD environment
type Environment interface {
	// Name returns the CI environment name
	Name() string

	// SetOutput sets an output variable
	SetOutput(key, value string) error

	// AnnotateError creates an error annotation
	AnnotateError(file string, line int, msg string) error

	// AnnotateWarning creates a warning annotation
	AnnotateWarning(file string, line int, msg string) error

	// IsPR returns true if running in a pull request
	IsPR() bool

	// PRNumber returns the pull request number (0 if not a PR)
	PRNumber() int
}

// Detector detects the current CI environment
type Detector struct{}

// NewDetector creates a new CI detector
func NewDetector() *Detector {
	return &Detector{}
}

// Detect identifies the current CI environment
func (d *Detector) Detect() Environment {
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		return NewGitHubActions()
	}
	if os.Getenv("GITLAB_CI") == "true" {
		return NewGitLabCI()
	}
	if os.Getenv("CIRCLECI") == "true" {
		return NewCircleCI()
	}
	return NewLocal()
}

// GitHubActions represents GitHub Actions environment
type GitHubActions struct{}

// NewGitHubActions creates a GitHub Actions environment
func NewGitHubActions() *GitHubActions {
	return &GitHubActions{}
}

func (g *GitHubActions) Name() string {
	return "GitHub Actions"
}

func (g *GitHubActions) SetOutput(key, value string) error {
	// GitHub Actions output format
	outputFile := os.Getenv("GITHUB_OUTPUT")
	if outputFile != "" {
		f, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = fmt.Fprintf(f, "%s=%s\n", key, value)
		return err
	}
	// Fallback to old format
	fmt.Printf("::set-output name=%s::%s\n", key, value)
	return nil
}

func (g *GitHubActions) AnnotateError(file string, line int, msg string) error {
	if file != "" && line > 0 {
		fmt.Printf("::error file=%s,line=%d::%s\n", file, line, msg)
	} else {
		fmt.Printf("::error::%s\n", msg)
	}
	return nil
}

func (g *GitHubActions) AnnotateWarning(file string, line int, msg string) error {
	if file != "" && line > 0 {
		fmt.Printf("::warning file=%s,line=%d::%s\n", file, line, msg)
	} else {
		fmt.Printf("::warning::%s\n", msg)
	}
	return nil
}

func (g *GitHubActions) IsPR() bool {
	return os.Getenv("GITHUB_EVENT_NAME") == "pull_request"
}

func (g *GitHubActions) PRNumber() int {
	if !g.IsPR() {
		return 0
	}
	prNum, _ := strconv.Atoi(os.Getenv("GITHUB_PR_NUMBER"))
	return prNum
}

// GitLabCI represents GitLab CI environment
type GitLabCI struct{}

// NewGitLabCI creates a GitLab CI environment
func NewGitLabCI() *GitLabCI {
	return &GitLabCI{}
}

func (g *GitLabCI) Name() string {
	return "GitLab CI"
}

func (g *GitLabCI) SetOutput(key, value string) error {
	// GitLab CI uses dotenv for artifacts
	f, err := os.OpenFile("build.env", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, "%s=%s\n", key, value)
	return err
}

func (g *GitLabCI) AnnotateError(file string, line int, msg string) error {
	// GitLab doesn't have native annotations, just print
	if file != "" && line > 0 {
		fmt.Printf("ERROR: %s:%d: %s\n", file, line, msg)
	} else {
		fmt.Printf("ERROR: %s\n", msg)
	}
	return nil
}

func (g *GitLabCI) AnnotateWarning(file string, line int, msg string) error {
	if file != "" && line > 0 {
		fmt.Printf("WARNING: %s:%d: %s\n", file, line, msg)
	} else {
		fmt.Printf("WARNING: %s\n", msg)
	}
	return nil
}

func (g *GitLabCI) IsPR() bool {
	return os.Getenv("CI_MERGE_REQUEST_ID") != ""
}

func (g *GitLabCI) PRNumber() int {
	if !g.IsPR() {
		return 0
	}
	prNum, _ := strconv.Atoi(os.Getenv("CI_MERGE_REQUEST_IID"))
	return prNum
}

// CircleCI represents CircleCI environment
type CircleCI struct{}

// NewCircleCI creates a CircleCI environment
func NewCircleCI() *CircleCI {
	return &CircleCI{}
}

func (c *CircleCI) Name() string {
	return "CircleCI"
}

func (c *CircleCI) SetOutput(key, value string) error {
	// CircleCI doesn't have native output mechanism
	fmt.Printf("%s=%s\n", key, value)
	return nil
}

func (c *CircleCI) AnnotateError(file string, line int, msg string) error {
	if file != "" && line > 0 {
		fmt.Printf("ERROR: %s:%d: %s\n", file, line, msg)
	} else {
		fmt.Printf("ERROR: %s\n", msg)
	}
	return nil
}

func (c *CircleCI) AnnotateWarning(file string, line int, msg string) error {
	if file != "" && line > 0 {
		fmt.Printf("WARNING: %s:%d: %s\n", file, line, msg)
	} else {
		fmt.Printf("WARNING: %s\n", msg)
	}
	return nil
}

func (c *CircleCI) IsPR() bool {
	return os.Getenv("CIRCLE_PULL_REQUEST") != ""
}

func (c *CircleCI) PRNumber() int {
	// CircleCI doesn't provide PR number directly
	return 0
}

// Local represents local development environment
type Local struct{}

// NewLocal creates a local environment
func NewLocal() *Local {
	return &Local{}
}

func (l *Local) Name() string {
	return "Local"
}

func (l *Local) SetOutput(key, value string) error {
	// In local mode, just print
	fmt.Printf("%s=%s\n", key, value)
	return nil
}

func (l *Local) AnnotateError(file string, line int, msg string) error {
	if file != "" && line > 0 {
		fmt.Printf("ERROR: %s:%d: %s\n", file, line, msg)
	} else {
		fmt.Printf("ERROR: %s\n", msg)
	}
	return nil
}

func (l *Local) AnnotateWarning(file string, line int, msg string) error {
	if file != "" && line > 0 {
		fmt.Printf("WARNING: %s:%d: %s\n", file, line, msg)
	} else {
		fmt.Printf("WARNING: %s\n", msg)
	}
	return nil
}

func (l *Local) IsPR() bool {
	return false
}

func (l *Local) PRNumber() int {
	return 0
}
