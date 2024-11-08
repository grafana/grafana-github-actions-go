package main

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
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

type ShellCommandRunner struct {
	Logger *slog.Logger
}

func NewShellCommandRunner(log *slog.Logger) *ShellCommandRunner {
	return &ShellCommandRunner{
		Logger: log,
	}
}

func (r *ShellCommandRunner) Run(ctx context.Context, command string, args ...string) (string, error) {
	var (
		stdout = bytes.NewBuffer(nil)
		stderr = bytes.NewBuffer(nil)
		cmdstr = strings.Join(append([]string{command}, args...), " ")
	)
	pwd, _ := os.Getwd()

	log := r.Logger.With("command", cmdstr, "wd", pwd)
	r.Logger.Debug("running command")

	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	log.Debug(fmt.Sprintf("stderr:\n%s", stderr.String()))
	log.Debug(fmt.Sprintf("stdout:\n%s", stdout.String()))

	err := cmd.Run()
	if err != nil {
		fmt.Errorf("error running command '%s': %w", cmdstr, err)
	}

	return strings.TrimSpace(stderr.String()), nil
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
