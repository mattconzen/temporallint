// Package hookcmd implements `temporallint hook install / uninstall /
// status`. It deliberately supports three integration paths so we
// don't conflict with whatever hook manager users already have:
//
//  - --manager=auto  (default) detect lefthook / pre-commit configs and
//    emit the right snippet; otherwise fall back to writing
//    .git/hooks/pre-commit directly.
//  - --manager=none  always write .git/hooks/pre-commit directly.
//  - --manager=lefthook / pre-commit  emit only the snippet; never
//    touch .git/hooks/.
//
// The DIY installer marks its hooks with a recognisable comment
// (`# managed-by: temporallint`) so subsequent install/uninstall calls
// are idempotent and don't clobber foreign hooks.
package hookcmd

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

// Main is the subcommand entry point. Returns the process exit code.
func Main(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage: temporallint hook install|uninstall|status [flags]")
		return 2
	}
	op := args[0]
	args = args[1:]

	fs := flag.NewFlagSet("temporallint hook "+op, flag.ContinueOnError)
	fs.SetOutput(stderr)
	mgr := fs.String("manager", "auto", "auto | none | lefthook | pre-commit")
	force := fs.Bool("force", false, "overwrite existing .git/hooks/pre-commit even if not managed by temporallint")
	dir := fs.String("repo", ".", "path to the git repository root")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	switch op {
	case "install":
		return runInstall(*dir, *mgr, *force, stdout, stderr)
	case "uninstall":
		return runUninstall(*dir, *mgr, stdout, stderr)
	case "status":
		return runStatus(*dir, *mgr, stdout, stderr)
	default:
		fmt.Fprintf(stderr, "unknown hook operation: %s\n", op)
		return 2
	}
}

func runInstall(repo, mgr string, force bool, stdout, stderr io.Writer) int {
	resolved, why := resolveManager(repo, mgr)
	switch resolved {
	case ManagerLefthook:
		fmt.Fprintf(stdout, "Detected %s; add this entry to lefthook.yml:\n\n%s\n", why, lefthookSnippet)
		fmt.Fprintln(stdout, "Then run: lefthook install")
		return 0
	case ManagerPrecommit:
		fmt.Fprintf(stdout, "Detected %s; add this entry to .pre-commit-config.yaml:\n\n%s\n", why, precommitSnippet)
		fmt.Fprintln(stdout, "Then run: pre-commit install")
		return 0
	case ManagerNone:
		path, err := installDIY(repo, force)
		if err != nil {
			fmt.Fprintf(stderr, "install: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "Installed pre-commit hook at %s\n", path)
		return 0
	default:
		fmt.Fprintf(stderr, "unknown manager: %s\n", mgr)
		return 2
	}
}

func runUninstall(repo, mgr string, stdout, stderr io.Writer) int {
	resolved, _ := resolveManager(repo, mgr)
	switch resolved {
	case ManagerLefthook:
		fmt.Fprintln(stdout, "Remove the temporallint entry from lefthook.yml and run: lefthook install")
		return 0
	case ManagerPrecommit:
		fmt.Fprintln(stdout, "Remove the temporallint repo from .pre-commit-config.yaml and run: pre-commit install")
		return 0
	case ManagerNone:
		removed, err := uninstallDIY(repo)
		if err != nil {
			if errors.Is(err, errForeignHook) {
				fmt.Fprintf(stderr, "uninstall: %v\n", err)
				return 1
			}
			fmt.Fprintf(stderr, "uninstall: %v\n", err)
			return 1
		}
		if removed {
			fmt.Fprintln(stdout, "Uninstalled pre-commit hook")
		} else {
			fmt.Fprintln(stdout, "No temporallint hook installed")
		}
		return 0
	default:
		fmt.Fprintf(stderr, "unknown manager: %s\n", mgr)
		return 2
	}
}

func runStatus(repo, mgr string, stdout, stderr io.Writer) int {
	resolved, why := resolveManager(repo, mgr)
	fmt.Fprintf(stdout, "manager: %s (%s)\n", resolved, why)
	if resolved == ManagerNone {
		state, err := statusDIY(repo)
		if err != nil {
			fmt.Fprintf(stderr, "status: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "diy hook: %s\n", state)
	}
	return 0
}

// Verify the file exists for the DIY path even when err is nil.
var _ = os.Stat
