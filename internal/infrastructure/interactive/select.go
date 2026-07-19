package interactive

import (
	"fmt"
	"os/exec"
	"strings"
)

// Selector abstracts fzf fuzzy selection for testing.
type Selector interface {
	// Select opens an fzf single-select prompt. Returns the chosen option.
	// ok=false means the user cancelled (ESC/Ctrl+C).
	Select(title string, options []string) (string, bool, error)

	// MultiSelect opens an fzf multi-select prompt. Returns chosen options.
	// Empty slice means nothing selected or cancelled.
	MultiSelect(title string, options []string) ([]string, error)
}

// FzfOptions configures the fzf TUI appearance.
type FzfOptions struct {
	Height  string // e.g. "40%"
	Reverse bool
}

// DefaultFzfOptions returns sensible defaults for the wizard.
func DefaultFzfOptions() FzfOptions {
	return FzfOptions{
		Height:  "40%",
		Reverse: true,
	}
}

// FzfSelector implements Selector by shelling out to the fzf binary.
type FzfSelector struct {
	path string
	opts FzfOptions
}

// NewFzfSelector looks up fzf in PATH and returns a selector.
// Returns an error if fzf is not installed.
func NewFzfSelector(opts FzfOptions) (*FzfSelector, error) {
	path, err := exec.LookPath("fzf")
	if err != nil {
		return nil, fmt.Errorf("fzf not found in PATH — instale com 'brew install fzf'")
	}
	return &FzfSelector{path: path, opts: opts}, nil
}

func (s *FzfSelector) Select(title string, options []string) (string, bool, error) {
	args := s.buildArgs(title, false)

	cmd := exec.Command(s.path, args...)
	cmd.Stdin = strings.NewReader(strings.Join(options, "\n"))

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code := exitErr.ExitCode()
			// 130 = SIGINT (Ctrl+C), 1 = ESC or empty match
			if code == 130 || code == 1 {
				return "", false, nil
			}
		}
		return "", false, fmt.Errorf("fzf failed: %w", err)
	}

	return strings.TrimSpace(string(output)), true, nil
}

func (s *FzfSelector) MultiSelect(title string, options []string) ([]string, error) {
	args := s.buildArgs(title, true)

	cmd := exec.Command(s.path, args...)
	cmd.Stdin = strings.NewReader(strings.Join(options, "\n"))

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code := exitErr.ExitCode()
			if code == 130 || code == 1 {
				return []string{}, nil
			}
		}
		return nil, fmt.Errorf("fzf failed: %w", err)
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		return []string{}, nil
	}
	return strings.Split(result, "\n"), nil
}

func (s *FzfSelector) buildArgs(title string, multi bool) []string {
	args := []string{
		"--prompt=" + title + "> ",
		"--no-info",
	}
	if multi {
		args = append(args, "--multi")
		args = append(args, "--header", "TAB seleciona/desmarca, ENTER confirma")
	}
	if s.opts.Height != "" {
		args = append(args, "--height="+s.opts.Height)
	}
	if s.opts.Reverse {
		args = append(args, "--reverse")
	}
	return args
}
