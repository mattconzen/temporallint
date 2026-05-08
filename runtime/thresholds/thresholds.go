// Package thresholds owns the default values for every runtime check and
// the YAML loader for --threshold-config overrides. Defaults are picked
// to be conservative: warn well below server hard limits, fail close to
// them so non-zero exit codes mean "you're about to break something".
package thresholds

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/mattconzen/monorepo/tools/temporallint/runtime"
)

// Defaults returns the baseline thresholds. Any zero field in the
// loaded YAML is replaced by the corresponding default so partial
// overrides work as expected.
func Defaults() runtime.Thresholds {
	return runtime.Thresholds{
		HistoryBytesWarn:       50 << 20,  // 50 MiB
		HistoryBytesFail:       200 << 20, // 200 MiB
		HistoryEventsWarn:      10_000,
		HistoryEventsFail:      50_000, // server hard cap is 51200
		IndividualPayloadWarn:  2 << 20, // 2 MiB
		IndividualPayloadFail:  4 << 20, // 4 MiB (server limit)
		NoWorkflowTimeoutAfter: 365 * 24 * time.Hour,
		SampleHistoryEvents:    256,
	}
}

// Load reads a YAML file and overlays it on Defaults(). Path may be
// empty, in which case Defaults() is returned.
func Load(path string) (runtime.Thresholds, error) {
	t := Defaults()
	if path == "" {
		return t, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return t, fmt.Errorf("read threshold config: %w", err)
	}
	var override runtime.Thresholds
	if err := yaml.Unmarshal(data, &override); err != nil {
		return t, fmt.Errorf("parse threshold config: %w", err)
	}
	merge(&t, override)
	return t, nil
}

func merge(dst *runtime.Thresholds, src runtime.Thresholds) {
	if src.HistoryBytesWarn > 0 {
		dst.HistoryBytesWarn = src.HistoryBytesWarn
	}
	if src.HistoryBytesFail > 0 {
		dst.HistoryBytesFail = src.HistoryBytesFail
	}
	if src.HistoryEventsWarn > 0 {
		dst.HistoryEventsWarn = src.HistoryEventsWarn
	}
	if src.HistoryEventsFail > 0 {
		dst.HistoryEventsFail = src.HistoryEventsFail
	}
	if src.IndividualPayloadWarn > 0 {
		dst.IndividualPayloadWarn = src.IndividualPayloadWarn
	}
	if src.IndividualPayloadFail > 0 {
		dst.IndividualPayloadFail = src.IndividualPayloadFail
	}
	if src.NoWorkflowTimeoutAfter > 0 {
		dst.NoWorkflowTimeoutAfter = src.NoWorkflowTimeoutAfter
	}
	if src.SampleHistoryEvents > 0 {
		dst.SampleHistoryEvents = src.SampleHistoryEvents
	}
}
