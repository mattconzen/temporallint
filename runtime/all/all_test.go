package all

import "testing"

// TestChecksUnique ensures no two checks share a name; collisions would
// silently mask one of them in CI output.
func TestChecksUnique(t *testing.T) {
	seen := map[string]struct{}{}
	for _, c := range Checks() {
		if _, dup := seen[c.Name()]; dup {
			t.Fatalf("duplicate check name: %s", c.Name())
		}
		seen[c.Name()] = struct{}{}
	}
	if len(Checks()) == 0 {
		t.Fatal("Checks() returned no checks")
	}
}

// TestChecksHaveURLs ensures every check links back to the source
// catalogue so users can read the prose explanation.
func TestChecksHaveURLs(t *testing.T) {
	for _, c := range Checks() {
		if c.URL() == "" {
			t.Errorf("check %q has empty URL", c.Name())
		}
		if c.Doc() == "" {
			t.Errorf("check %q has empty Doc", c.Name())
		}
	}
}
