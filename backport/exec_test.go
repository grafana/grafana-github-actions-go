package main

import (
	"context"
	"strings"
	"sync"
)

// mockRunner is a CommandRunner stub used by tests. It records every command it sees and
// returns canned stdout from `Outputs` when present (lookups not in the map return "").
type mockRunner struct {
	Commands []string
	Outputs  map[string]string
	mtx      sync.Mutex
}

func newMockRunner() *mockRunner {
	return &mockRunner{Commands: []string{}}
}

func (m *mockRunner) Run(ctx context.Context, command string, args ...string) (string, error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	cmd := strings.Join(append([]string{command}, args...), " ")
	m.Commands = append(m.Commands, cmd)
	if out, ok := m.Outputs[cmd]; ok {
		return out, nil
	}
	return "", nil
}

// errorRunner is a CommandRunner stub that returns a canned error for specific commands. It also
// records the commands it saw via an embedded mockRunner so tests can assert on the sequence.
type errorRunner struct {
	Errors  map[string]error
	History *mockRunner
}

func newErrorRunner(errors map[string]error) *errorRunner {
	return &errorRunner{Errors: errors, History: newMockRunner()}
}

func (r *errorRunner) Run(ctx context.Context, command string, args ...string) (string, error) {
	cmd := strings.Join(append([]string{command}, args...), " ")
	out, _ := r.History.Run(ctx, command, args...)
	if err, ok := r.Errors[cmd]; ok {
		return "", err
	}
	return out, nil
}
