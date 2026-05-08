package hookcmd

import (
	"os"
	"path/filepath"
)

// Manager identifies which hook integration path is in effect.
type Manager string

const (
	ManagerAuto      Manager = "auto"
	ManagerNone      Manager = "none"
	ManagerLefthook  Manager = "lefthook"
	ManagerPrecommit Manager = "pre-commit"
	ManagerUnknown   Manager = "unknown"
)

// resolveManager picks a Manager based on the user's flag and the
// repo's existing config files. Returns the chosen Manager and a short
// human reason explaining why.
func resolveManager(repo, requested string) (Manager, string) {
	switch Manager(requested) {
	case ManagerNone:
		return ManagerNone, "explicitly requested"
	case ManagerLefthook:
		return ManagerLefthook, "explicitly requested"
	case ManagerPrecommit:
		return ManagerPrecommit, "explicitly requested"
	case ManagerAuto, "":
		// Detection order: lefthook first (Go-native, more common in
		// fresh repos), then pre-commit, then DIY.
		if fileExists(filepath.Join(repo, "lefthook.yml")) || fileExists(filepath.Join(repo, "lefthook.yaml")) {
			return ManagerLefthook, "found lefthook config"
		}
		if fileExists(filepath.Join(repo, ".pre-commit-config.yaml")) {
			return ManagerPrecommit, "found .pre-commit-config.yaml"
		}
		return ManagerNone, "no manager detected; using direct .git/hooks install"
	}
	return ManagerUnknown, "unknown manager"
}

func fileExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}

const lefthookSnippet = `pre-commit:
  commands:
    temporallint:
      run: go run ./tools/temporallint/cmd/temporallint ./...
`

const precommitSnippet = `repos:
  - repo: local
    hooks:
      - id: temporallint
        name: temporallint
        entry: go run ./tools/temporallint/cmd/temporallint ./...
        language: system
        pass_filenames: false
`
