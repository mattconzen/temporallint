package all

import (
	"sort"
	"testing"
)

// TestAnalyzersUnique ensures no two analyzers share a name. Duplicate
// names would silently mask diagnostics from one of them at the
// multichecker layer.
func TestAnalyzersUnique(t *testing.T) {
	seen := map[string]struct{}{}
	for _, a := range Analyzers() {
		if _, dup := seen[a.Name]; dup {
			t.Fatalf("duplicate analyzer name: %s", a.Name)
		}
		seen[a.Name] = struct{}{}
	}
	if len(Analyzers()) == 0 {
		t.Fatal("Analyzers() returned no live analyzers")
	}
}

// TestRegistryUniqueNames covers Implemented + Planned. The generated
// RULES.md table is keyed by Name; duplicates would corrupt it.
func TestRegistryUniqueNames(t *testing.T) {
	names := map[string]struct{}{}
	for _, e := range Registry() {
		if _, dup := names[e.Name]; dup {
			t.Fatalf("duplicate registry entry name: %s", e.Name)
		}
		names[e.Name] = struct{}{}
	}
}

// TestImplementedHaveAnalyzer guards against accidentally marking a rule
// Implemented without wiring up the Analyzer.
func TestImplementedHaveAnalyzer(t *testing.T) {
	for _, e := range Registry() {
		if e.Status == StatusImplemented && e.Analyzer == nil {
			t.Errorf("rule %q is marked Implemented but has no Analyzer", e.Name)
		}
	}
}

// TestRegistrySorted is a smoke check that Registry() returns a stable
// order — used by gen-rules to keep RULES.md diffs minimal.
func TestRegistrySorted(t *testing.T) {
	got := Registry()
	want := append([]Entry(nil), got...)
	sort.SliceStable(want, func(i, j int) bool {
		if rank(want[i].Status) != rank(want[j].Status) {
			return rank(want[i].Status) < rank(want[j].Status)
		}
		return want[i].Name < want[j].Name
	})
	for i := range got {
		if got[i].Name != want[i].Name {
			t.Fatalf("Registry() not sorted; first divergence at %d: got %q want %q", i, got[i].Name, want[i].Name)
		}
	}
}
