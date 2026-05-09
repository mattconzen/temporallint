package violation

import (
	"crypto/sha256"
	"encoding/json"
	"regexp"

	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	_ = sha256.Sum256([]byte("hello")) // want `sha256.Sum256 in workflow code`
	_, _ = json.Marshal(map[string]int{"a": 1}) // want `encoding/json.Marshal in workflow code`
	_ = regexp.MustCompile(`^foo$`)             // want `regexp.MustCompile in workflow code`
	return nil
}
