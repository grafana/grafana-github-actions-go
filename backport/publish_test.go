package main

import (
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/grafana/grafana-github-actions-go/pkg/ghgql"
	"github.com/grafana/grafana-github-actions-go/pkg/ghutil"
	"github.com/stretchr/testify/require"
)

func TestSplitMessage(t *testing.T) {
	cases := []struct {
		in           string
		wantHeadline string
		wantBody     string
	}{
		{"Just a subject", "Just a subject", ""},
		{"Subject\n\nBody paragraph", "Subject", "Body paragraph"},
		{"Subject\n\nBody paragraph 1\n\nBody paragraph 2", "Subject", "Body paragraph 1\n\nBody paragraph 2"},
		{"Subject\nNo blank line", "Subject", "No blank line"},
		{"Subject\n\nBody\n\n(cherry picked from commit abc)\n", "Subject", "Body\n\n(cherry picked from commit abc)"},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			h, b := splitMessage(tc.in)
			require.Equal(t, tc.wantHeadline, h)
			require.Equal(t, tc.wantBody, b)
		})
	}
}

func TestAppendCoAuthor(t *testing.T) {
	t.Run("appends trailer when both name and email are set", func(t *testing.T) {
		out := appendCoAuthor("Subject\n\nBody", "Alice Example", "alice@example.com")
		require.Equal(t, "Subject\n\nBody\n\nCo-authored-by: Alice Example <alice@example.com>", out)
	})
	t.Run("omits trailer when name is empty", func(t *testing.T) {
		out := appendCoAuthor("Subject", "", "alice@example.com")
		require.Equal(t, "Subject", out)
	})
	t.Run("omits trailer when email is empty", func(t *testing.T) {
		out := appendCoAuthor("Subject", "Alice", "")
		require.Equal(t, "Subject", out)
	})
	t.Run("strips trailing newlines from message before appending", func(t *testing.T) {
		out := appendCoAuthor("Subject\n\n", "Alice", "a@x")
		require.Equal(t, "Subject\n\nCo-authored-by: Alice <a@x>", out)
	})
}

// publisherTestRepo isolates both the Git objects and the working directory.
func publisherTestRepo(t *testing.T) func(...string) string {
	t.Helper()
	t.Chdir(t.TempDir())
	git := func(args ...string) string {
		t.Helper()
		out, err := exec.CommandContext(t.Context(), "git", args...).CombinedOutput()
		require.NoError(t, err, "%s", out)
		return strings.TrimSpace(string(out))
	}
	git("init")
	git("config", "user.name", "Example")
	git("config", "user.email", "example@example.com")
	// Avoid signing-key requirements and ensure mode-only changes are visible.
	git("config", "commit.gpgsign", "false")
	git("config", "core.filemode", "true")
	return git
}

func publisherTestRunner() *ShellCommandRunner {
	return NewShellCommandRunner(slog.New(slog.NewTextHandler(io.Discard, nil)))
}

type publisherModeCase struct {
	name   string
	change func(*testing.T)
}

func publisherModeCases() []publisherModeCase {
	return []publisherModeCase{
		{"added executable", func(t *testing.T) {
			require.NoError(t, os.WriteFile("added", []byte("new"), 0o755))
		}},
		{"modified executable", func(t *testing.T) {
			require.NoError(t, os.WriteFile("executable", []byte("new"), 0o755))
		}},
		{"added symlink", func(t *testing.T) {
			require.NoError(t, os.Symlink("missing", "added"))
		}},
		{"modified symlink", func(t *testing.T) {
			require.NoError(t, os.Remove("link"))
			require.NoError(t, os.Symlink("missing", "link"))
		}},
		{"regular to executable", func(t *testing.T) {
			require.NoError(t, os.Chmod("regular", 0o755))
		}},
		{"executable to regular", func(t *testing.T) {
			require.NoError(t, os.Chmod("executable", 0o644))
		}},
		{"regular to symlink", func(t *testing.T) {
			require.NoError(t, os.Remove("regular"))
			require.NoError(t, os.Symlink("missing", "regular"))
		}},
		{"symlink to regular", func(t *testing.T) {
			require.NoError(t, os.Remove("link"))
			require.NoError(t, os.WriteFile("link", []byte("new"), 0o644))
		}},
	}
}

type publisherInvalidDiffCase struct{ name, raw, message string }

func publisherInvalidDiffCases() []publisherInvalidDiffCase {
	return []publisherInvalidDiffCase{
		{"missing terminator", "malformed", "malformed raw diff"},
		{"missing path", ":000000 100644 old new A\x00", "malformed raw diff"},
		{"empty path", ":000000 100644 old new A\x00\x00", "malformed raw diff entry"},
		{"missing fields", "A\x00path\x00", "malformed raw diff entry"},
		{"unsupported status", ":100644 100644 old new X\x00path\x00", "manual signed backport required"},
		{"missing blob", ":000000 100644 0000000000000000000000000000000000000000 ffffffffffffffffffffffffffffffffffffffff A\x00missing\x00", "reading Git blob"},
	}
}

// publisherTestUnsupportedMode returns the base, with origin/base at that commit
// and HEAD containing the mixed batch.
func publisherTestUnsupportedMode(t *testing.T, change func(*testing.T)) string {
	t.Helper()
	git := publisherTestRepo(t)
	require.NoError(t, os.WriteFile("regular", []byte("old"), 0o644))
	require.NoError(t, os.WriteFile("executable", []byte("old"), 0o755))
	require.NoError(t, os.Symlink("regular", "link"))
	git("add", ".")
	git("commit", "-m", "base")
	base := git("rev-parse", "HEAD")
	git("update-ref", "refs/remotes/origin/base", base)
	// A supported change comes first: the entire batch must fail before publication.
	require.NoError(t, os.WriteFile("a-regular", []byte("supported"), 0o644))
	change(t)
	git("add", "-A")
	git("commit", "-m", "unsupported change")
	return base
}

// publisherTestSubmodule returns the base, with origin/base at that commit
// and HEAD containing the requested gitlink operation.
func publisherTestSubmodule(t *testing.T, operation string) string {
	t.Helper()
	git := publisherTestRepo(t)
	git("commit", "--allow-empty", "-m", "base")
	base := git("rev-parse", "HEAD")
	// Construct gitlinks (160000) directly in the index using commit objects.
	git("update-index", "--add", "--cacheinfo", "160000", base, "module")
	git("commit", "-m", "add module")
	switch operation {
	case "addition":
		// HEAD already contains the added gitlink.
	case "modification":
		base = git("rev-parse", "HEAD")
		git("update-index", "--cacheinfo", "160000", base, "module")
		git("commit", "-m", "update module")
	case "deletion":
		base = git("rev-parse", "HEAD")
		git("update-index", "--force-remove", "module")
		git("commit", "-m", "delete module")
	default:
		t.Fatalf("unknown submodule operation %q", operation)
	}
	git("update-ref", "refs/remotes/origin/base", base)
	return base
}

func TestBuildFileChanges(t *testing.T) {
	t.Run("committed contents", func(t *testing.T) {
		git := publisherTestRepo(t)
		require.NoError(t, os.WriteFile("modified", []byte("old"), 0o644))
		require.NoError(t, os.WriteFile("deleted", []byte("old"), 0o644))
		git("add", ".")
		git("commit", "-m", "base")
		base := git("rev-parse", "HEAD")

		// Binary bytes and surrounding whitespace must survive unchanged; the newline
		// in the path exercises Git's NUL-delimited filename format.
		content := []byte(" \t\x00binary\xff\n\n")
		path := "added name\n.txt"
		require.NoError(t, os.WriteFile(path, content, 0o644))
		require.NoError(t, os.WriteFile("empty", nil, 0o644))
		require.NoError(t, os.WriteFile("modified", nil, 0o644))
		require.NoError(t, os.Remove("deleted"))
		git("add", "-A")
		git("commit", "-m", "mixed changes")

		// Publication must use committed bytes, even when the working tree is stale or absent.
		require.NoError(t, os.Remove(path))
		require.NoError(t, os.Remove("empty"))
		require.NoError(t, os.WriteFile("modified", []byte("uncommitted"), 0o644))
		changes, err := buildFileChanges(t.Context(), publisherTestRunner(), base, "HEAD")
		require.NoError(t, err)
		require.Equal(t, []ghgql.FileChange{
			{Path: path, Contents: content},
			{Path: "deleted", Delete: true},
			{Path: "empty", Contents: []byte{}},
			{Path: "modified", Contents: []byte{}},
		}, changes)
	})

	t.Run("empty diff", func(t *testing.T) {
		git := publisherTestRepo(t)
		git("commit", "--allow-empty", "-m", "base")
		changes, err := buildFileChanges(t.Context(), publisherTestRunner(), "HEAD", "HEAD")
		require.NoError(t, err)
		require.Empty(t, changes)
	})

	t.Run("unsupported modes", func(t *testing.T) {
		// Only 100644 writes are supported; the API cannot encode other modes or type changes.
		for _, tc := range publisherModeCases() {
			t.Run(tc.name, func(t *testing.T) {
				base := publisherTestUnsupportedMode(t, tc.change)
				changes, err := buildFileChanges(t.Context(), publisherTestRunner(), base, "HEAD")
				require.ErrorContains(t, err, "manual signed backport required")
				require.Nil(t, changes)
			})
		}
	})

	t.Run("submodules", func(t *testing.T) {
		for _, operation := range []string{"addition", "modification"} {
			t.Run(operation, func(t *testing.T) {
				base := publisherTestSubmodule(t, operation)
				_, err := buildFileChanges(t.Context(), publisherTestRunner(), base, "HEAD")
				require.ErrorContains(t, err, "manual signed backport required")
			})
		}
		t.Run("deletion", func(t *testing.T) {
			base := publisherTestSubmodule(t, "deletion")
			// Removing a gitlink needs only its path, so deletion is representable.
			changes, err := buildFileChanges(t.Context(), publisherTestRunner(), base, "HEAD")
			require.NoError(t, err)
			require.Equal(t, []ghgql.FileChange{{Path: "module", Delete: true}}, changes)
		})
	})

	t.Run("deletes non-regular files", func(t *testing.T) {
		// The API can delete these entries by path without encoding their original mode.
		git := publisherTestRepo(t)
		require.NoError(t, os.WriteFile("executable", []byte("old"), 0o755))
		require.NoError(t, os.Symlink("missing", "link"))
		git("add", ".")
		git("commit", "-m", "base")
		base := git("rev-parse", "HEAD")
		git("rm", "executable", "link")
		git("commit", "-m", "delete")
		changes, err := buildFileChanges(t.Context(), publisherTestRunner(), base, "HEAD")
		require.NoError(t, err)
		require.Equal(t, []ghgql.FileChange{{Path: "executable", Delete: true}, {Path: "link", Delete: true}}, changes)
	})

	t.Run("malformed diff or missing blob", func(t *testing.T) {
		for _, tc := range publisherInvalidDiffCases() {
			t.Run(tc.name, func(t *testing.T) {
				publisherTestRepo(t)
				runner := newMockRunner()
				runner.Outputs = map[string]string{"git diff --no-renames --raw --no-abbrev -z base HEAD": tc.raw}
				_, err := buildFileChanges(t.Context(), runner, "base", "HEAD")
				require.ErrorContains(t, err, tc.message)
			})
		}
	})

	t.Run("preserves diff order", func(t *testing.T) {
		runner := newMockRunner()
		// Use deliberately unsorted paths to detect accidental reordering.
		runner.Outputs = map[string]string{"git diff --no-renames --raw --no-abbrev -z base HEAD": ":100644 000000 old zero D\x00z\x00:100644 000000 old zero D\x00a\x00"}
		changes, err := buildFileChanges(t.Context(), runner, "base", "HEAD")
		require.NoError(t, err)
		require.Equal(t, []ghgql.FileChange{{Path: "z", Delete: true}, {Path: "a", Delete: true}}, changes)
	})

	t.Run("propagates diff error", func(t *testing.T) {
		runner := newErrorRunner(map[string]error{"git diff --no-renames --raw --no-abbrev -z base HEAD": os.ErrInvalid})
		_, err := buildFileChanges(t.Context(), runner, "base", "HEAD")
		require.ErrorIs(t, err, os.ErrInvalid)
	})
}

func TestPublishViaAPI(t *testing.T) {
	// Each case must fail before branch or commit creation; nil publication
	// clients make any attempted remote write fail the test.
	t.Run("empty diff", func(t *testing.T) {
		git := publisherTestRepo(t)
		git("commit", "--allow-empty", "-m", "base")
		git("update-ref", "refs/remotes/origin/base", "HEAD")
		err := PublishViaAPI(t.Context(), publisherTestRunner(), nil, nil, "target", BackportOpts{Target: ghutil.Branch{Name: "base"}})
		require.EqualError(t, err, "no file changes detected between origin/base and HEAD after cherry-pick; refusing to create an empty commit")
	})

	t.Run("unsupported modes", func(t *testing.T) {
		for _, tc := range publisherModeCases() {
			t.Run(tc.name, func(t *testing.T) {
				publisherTestUnsupportedMode(t, tc.change)
				err := PublishViaAPI(t.Context(), publisherTestRunner(), nil, nil, "target", BackportOpts{Target: ghutil.Branch{Name: "base"}})
				require.ErrorContains(t, err, "manual signed backport required")
			})
		}
	})

	t.Run("submodules", func(t *testing.T) {
		for _, operation := range []string{"addition", "modification"} {
			t.Run(operation, func(t *testing.T) {
				publisherTestSubmodule(t, operation)
				err := PublishViaAPI(t.Context(), publisherTestRunner(), nil, nil, "target", BackportOpts{Target: ghutil.Branch{Name: "base"}})
				require.ErrorContains(t, err, "manual signed backport required")
			})
		}
	})

	t.Run("malformed diff or missing blob", func(t *testing.T) {
		for _, tc := range publisherInvalidDiffCases() {
			t.Run(tc.name, func(t *testing.T) {
				publisherTestRepo(t)
				runner := newMockRunner()
				runner.Outputs = map[string]string{"git diff --no-renames --raw --no-abbrev -z base HEAD": tc.raw, "git rev-parse origin/base": "base"}
				err := PublishViaAPI(t.Context(), runner, nil, nil, "target", BackportOpts{Target: ghutil.Branch{Name: "base"}})
				require.ErrorContains(t, err, tc.message)
			})
		}
	})
}
