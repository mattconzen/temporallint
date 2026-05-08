package hookcmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// hookMarker is written into every DIY-installed hook and looked up
// by uninstall to confirm we own the file. Foreign hooks are left
// alone unless --force is set.
const hookMarker = "# managed-by: temporallint"

const hookScript = `#!/bin/sh
` + hookMarker + `
# Auto-installed by ` + "`" + `temporallint hook install` + "`" + `. To remove, run
# ` + "`" + `temporallint hook uninstall` + "`" + ` or delete this file.
exec go run ./tools/temporallint/cmd/temporallint ./...
`

var errForeignHook = errors.New("existing pre-commit hook is not managed by temporallint; rerun with --force to overwrite")

// installDIY writes the pre-commit hook script into .git/hooks/. If a
// hook already exists and isn't owned by temporallint, returns
// errForeignHook unless force is set.
func installDIY(repo string, force bool) (string, error) {
	hookDir, err := gitHooksDir(repo)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(hookDir, 0o755); err != nil {
		return "", fmt.Errorf("ensure hooks dir: %w", err)
	}
	path := filepath.Join(hookDir, "pre-commit")
	if existing, err := os.ReadFile(path); err == nil {
		if !isOurs(existing) && !force {
			return path, errForeignHook
		}
	} else if !os.IsNotExist(err) {
		return path, fmt.Errorf("read existing hook: %w", err)
	}
	if err := os.WriteFile(path, []byte(hookScript), 0o755); err != nil {
		return path, fmt.Errorf("write hook: %w", err)
	}
	return path, nil
}

// uninstallDIY removes the pre-commit hook iff temporallint installed
// it. Returns (true, nil) if a hook was removed, (false, nil) if no
// hook was present, errForeignHook if the hook is foreign.
func uninstallDIY(repo string) (bool, error) {
	hookDir, err := gitHooksDir(repo)
	if err != nil {
		return false, err
	}
	path := filepath.Join(hookDir, "pre-commit")
	existing, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("read existing hook: %w", err)
	}
	if !isOurs(existing) {
		return false, errForeignHook
	}
	if err := os.Remove(path); err != nil {
		return false, fmt.Errorf("remove hook: %w", err)
	}
	return true, nil
}

// statusDIY returns a human-readable status string describing the
// .git/hooks/pre-commit state.
func statusDIY(repo string) (string, error) {
	hookDir, err := gitHooksDir(repo)
	if err != nil {
		return "", err
	}
	path := filepath.Join(hookDir, "pre-commit")
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "not installed", nil
		}
		return "", err
	}
	if isOurs(b) {
		return "installed (temporallint-managed)", nil
	}
	return "exists but not managed by temporallint", nil
}

func isOurs(content []byte) bool {
	return strings.Contains(string(content), hookMarker)
}

// gitHooksDir resolves the hook directory for the given repo path.
// Honours `core.hooksPath` if set (lefthook and others use it). Falls
// back to <repo>/.git/hooks for ordinary repos and worktrees.
func gitHooksDir(repo string) (string, error) {
	gitDir := filepath.Join(repo, ".git")
	st, err := os.Stat(gitDir)
	if err != nil {
		return "", fmt.Errorf("not a git repo (%s): %w", repo, err)
	}
	if !st.IsDir() {
		// .git is a file → worktree pointer. Read it for the gitdir.
		b, err := os.ReadFile(gitDir)
		if err != nil {
			return "", fmt.Errorf("read .git pointer: %w", err)
		}
		line := strings.TrimSpace(strings.TrimPrefix(string(b), "gitdir: "))
		if !filepath.IsAbs(line) {
			line = filepath.Join(repo, line)
		}
		gitDir = line
	}
	return filepath.Join(gitDir, "hooks"), nil
}
