package main

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
)

type CommandRunner interface {
	Run(ctx context.Context, command string, args ...string) (string, error)
}

type NoOpRunner struct {
	Commands []string
	mtx      sync.Mutex
}

func NewNoOpRunner() *NoOpRunner {
	return &NoOpRunner{
		Commands: []string{},
		mtx:      sync.Mutex{},
	}
}

func (n *NoOpRunner) Run(ctx context.Context, command string, args ...string) (string, error) {
	n.mtx.Lock()
	cmd := strings.Join(append([]string{command}, args...), " ")
	n.Commands = append(n.Commands, cmd)
	defer n.mtx.Unlock()

	return "", nil
}

type ShellCommandRunner struct{}

func NewShellCommandRunner() *ShellCommandRunner {
	return &ShellCommandRunner{}
}

func (r *ShellCommandRunner) Run(ctx context.Context, command string, args ...string) (string, error) {
	var (
		stderr = bytes.NewBuffer(nil)
	)
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Stderr = stderr

	err := cmd.Run()
	cmdstr := strings.Join(append([]string{command}, args...), " ")

	return strings.TrimSpace(stderr.String()), fmt.Errorf("error running command '%s': %w", cmdstr, err)
}

type ErrorRunner struct {
	Commands map[string]error
}

func NewErrorRunner(errors map[string]error) *ErrorRunner {
	return &ErrorRunner{
		Commands: errors,
	}
}

func (r *ErrorRunner) Run(ctx context.Context, command string, args ...string) (string, error) {
	cmd := strings.Join(append([]string{command}, args...), " ")
	if err, ok := r.Commands[cmd]; ok {
		return "", err
	}

	return "", nil
}
