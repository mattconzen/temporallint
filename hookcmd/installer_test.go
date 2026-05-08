package hookcmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// freshRepo creates a temporary directory with a minimal .git
// directory so the installer's gitHooksDir() succeeds.
func freshRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git", "hooks"), 0o755); err != nil {
		t.Fatalf("init repo: %v", err)
	}
	return dir
}

func TestResolveManagerLefthook(t *testing.T) {
	dir := freshRepo(t)
	if err := os.WriteFile(filepath.Join(dir, "lefthook.yml"), []byte("pre-commit:\n"), 0o644); err != nil {
		t.Fatalf("write lefthook.yml: %v", err)
	}
	got, why := resolveManager(dir, "auto")
	if got != ManagerLefthook {
		t.Fatalf("want lefthook, got %s (%s)", got, why)
	}
}

func TestResolveManagerPrecommit(t *testing.T) {
	dir := freshRepo(t)
	if err := os.WriteFile(filepath.Join(dir, ".pre-commit-config.yaml"), []byte("repos: []\n"), 0o644); err != nil {
		t.Fatalf("write pre-commit config: %v", err)
	}
	got, why := resolveManager(dir, "auto")
	if got != ManagerPrecommit {
		t.Fatalf("want pre-commit, got %s (%s)", got, why)
	}
}

func TestResolveManagerFallbackToNone(t *testing.T) {
	dir := freshRepo(t)
	got, why := resolveManager(dir, "auto")
	if got != ManagerNone {
		t.Fatalf("want none, got %s (%s)", got, why)
	}
}

func TestInstallDIYHappyPath(t *testing.T) {
	dir := freshRepo(t)
	path, err := installDIY(dir, false)
	if err != nil {
		t.Fatalf("install: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read hook: %v", err)
	}
	if !strings.Contains(string(got), hookMarker) {
		t.Fatalf("installed hook missing marker; content:\n%s", got)
	}
	st, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat hook: %v", err)
	}
	if st.Mode()&0o111 == 0 {
		t.Fatalf("hook is not executable: %v", st.Mode())
	}
}

func TestInstallDIYIdempotent(t *testing.T) {
	dir := freshRepo(t)
	if _, err := installDIY(dir, false); err != nil {
		t.Fatalf("first install: %v", err)
	}
	if _, err := installDIY(dir, false); err != nil {
		t.Fatalf("second install: %v", err)
	}
}

func TestInstallDIYRefusesForeignHook(t *testing.T) {
	dir := freshRepo(t)
	hookPath := filepath.Join(dir, ".git", "hooks", "pre-commit")
	if err := os.WriteFile(hookPath, []byte("#!/bin/sh\necho foreign\n"), 0o755); err != nil {
		t.Fatalf("write foreign hook: %v", err)
	}
	_, err := installDIY(dir, false)
	if err != errForeignHook {
		t.Fatalf("want errForeignHook, got %v", err)
	}
	// --force must overwrite.
	if _, err := installDIY(dir, true); err != nil {
		t.Fatalf("--force install: %v", err)
	}
	got, _ := os.ReadFile(hookPath)
	if !strings.Contains(string(got), hookMarker) {
		t.Fatalf("--force did not overwrite; content:\n%s", got)
	}
}

func TestUninstallDIY(t *testing.T) {
	dir := freshRepo(t)
	if _, err := installDIY(dir, false); err != nil {
		t.Fatalf("install: %v", err)
	}
	removed, err := uninstallDIY(dir)
	if err != nil {
		t.Fatalf("uninstall: %v", err)
	}
	if !removed {
		t.Fatal("expected removed=true")
	}
	if _, err := os.Stat(filepath.Join(dir, ".git", "hooks", "pre-commit")); !os.IsNotExist(err) {
		t.Fatalf("hook should be gone, got err=%v", err)
	}
}

func TestUninstallDIYAbsent(t *testing.T) {
	dir := freshRepo(t)
	removed, err := uninstallDIY(dir)
	if err != nil {
		t.Fatalf("uninstall: %v", err)
	}
	if removed {
		t.Fatal("expected removed=false when no hook present")
	}
}

func TestUninstallDIYRefusesForeign(t *testing.T) {
	dir := freshRepo(t)
	hookPath := filepath.Join(dir, ".git", "hooks", "pre-commit")
	if err := os.WriteFile(hookPath, []byte("#!/bin/sh\necho foreign\n"), 0o755); err != nil {
		t.Fatalf("write foreign hook: %v", err)
	}
	if _, err := uninstallDIY(dir); err != errForeignHook {
		t.Fatalf("want errForeignHook, got %v", err)
	}
	// And the foreign hook must remain untouched.
	got, _ := os.ReadFile(hookPath)
	if !strings.Contains(string(got), "echo foreign") {
		t.Fatalf("foreign hook was modified; content:\n%s", got)
	}
}

func TestStatusDIY(t *testing.T) {
	dir := freshRepo(t)
	got, err := statusDIY(dir)
	if err != nil || got != "not installed" {
		t.Fatalf("want 'not installed', got %q err=%v", got, err)
	}
	if _, err := installDIY(dir, false); err != nil {
		t.Fatalf("install: %v", err)
	}
	got, err = statusDIY(dir)
	if err != nil || !strings.Contains(got, "temporallint-managed") {
		t.Fatalf("want temporallint-managed, got %q err=%v", got, err)
	}
}

func TestGitHooksDirHandlesWorktreePointer(t *testing.T) {
	dir := t.TempDir()
	// Simulate a git worktree where .git is a file pointing at a real gitdir.
	realGitDir := filepath.Join(dir, "gitdir")
	if err := os.MkdirAll(filepath.Join(realGitDir, "hooks"), 0o755); err != nil {
		t.Fatalf("mkdir gitdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".git"), []byte("gitdir: "+realGitDir+"\n"), 0o644); err != nil {
		t.Fatalf("write .git pointer: %v", err)
	}
	got, err := gitHooksDir(dir)
	if err != nil {
		t.Fatalf("gitHooksDir: %v", err)
	}
	if got != filepath.Join(realGitDir, "hooks") {
		t.Fatalf("got %q want %q", got, filepath.Join(realGitDir, "hooks"))
	}
}
