package interactive

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// Prompter abstracts text and confirmation prompts for testing.
type Prompter interface {
	Text(prompt, defaultValue string) (string, error)
	Confirm(prompt string, defaultYes bool) (bool, error)
}

// StdioPrompter reads from an io.Reader and writes to an io.Writer.
type StdioPrompter struct {
	reader *bufio.Reader
	writer io.Writer
}

// NewStdioPrompter creates a prompter using os.Stdin and os.Stdout.
func NewStdioPrompter() *StdioPrompter {
	return &StdioPrompter{
		reader: bufio.NewReader(os.Stdin),
		writer: os.Stdout,
	}
}

// NewStdioPrompterWithIO creates a prompter with custom I/O (for testing).
func NewStdioPrompterWithIO(r io.Reader, w io.Writer) *StdioPrompter {
	return &StdioPrompter{
		reader: bufio.NewReader(r),
		writer: w,
	}
}

func (p *StdioPrompter) Text(prompt, defaultValue string) (string, error) {
	if defaultValue != "" {
		fmt.Fprintf(p.writer, "%s [%s]: ", prompt, defaultValue)
	} else {
		fmt.Fprintf(p.writer, "%s: ", prompt)
	}

	input, err := p.reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue, nil
	}
	return input, nil
}

func (p *StdioPrompter) Confirm(prompt string, defaultYes bool) (bool, error) {
	defaultHint := "[Y/n]"
	if !defaultYes {
		defaultHint = "[y/N]"
	}

	fmt.Fprintf(p.writer, "%s %s: ", prompt, defaultHint)

	input, err := p.reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	input = strings.TrimSpace(strings.ToLower(input))
	if input == "" {
		return defaultYes, nil
	}

	switch input {
	case "y", "yes", "s", "sim":
		return true, nil
	default:
		return false, nil
	}
}
