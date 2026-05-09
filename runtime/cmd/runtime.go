// Package cmd implements `temporallint runtime`. It parses flags, dials
// a Temporal server, fans out every registered runtime check in
// parallel, collects findings, and writes them as text or JSON.
//
// Wired into the main binary at tools/temporallint/cmd/temporallint/main.go
// via argv dispatch: `temporallint runtime ...` lands here, anything
// else falls through to the static-analysis multichecker.
package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"go.temporal.io/sdk/client"

	"github.com/mattconzen/temporallint/runtime"
	runtimeall "github.com/mattconzen/temporallint/runtime/all"
	"github.com/mattconzen/temporallint/runtime/internal/temporaladapter"
	"github.com/mattconzen/temporallint/runtime/thresholds"
)

// Main is the entry point. It returns the process exit code rather than
// calling os.Exit so it can be tested.
func Main(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("temporallint runtime", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var (
		address    = fs.String("address", "localhost:7233", "Temporal frontend gRPC address")
		namespace  = fs.String("namespace", "default", "Temporal namespace")
		apiKey     = fs.String("api-key", os.Getenv("TEMPORAL_API_KEY"), "Temporal Cloud / gateway API key (defaults to TEMPORAL_API_KEY env)")
		taskQueue  = fs.String("task-queue", "", "optional task-queue filter")
		since      = fs.Duration("since", 24*time.Hour, "list workflows started within this duration")
		thresholdF = fs.String("threshold-config", "", "path to YAML threshold overrides")
		output     = fs.String("output", "text", "output format: text or json")
		failOn     = fs.String("fail-on", "fail", "minimum severity that produces non-zero exit: ok | warn | fail")
	)
	if err := fs.Parse(args); err != nil {
		return 2
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	thresh, err := thresholds.Load(*thresholdF)
	if err != nil {
		fmt.Fprintf(stderr, "load threshold config: %v\n", err)
		return 2
	}

	c, err := dial(*address, *namespace, *apiKey)
	if err != nil {
		fmt.Fprintf(stderr, "dial temporal: %v\n", err)
		return 2
	}
	defer c.Close()

	api := temporaladapter.New(c, *namespace)
	deps := runtime.Deps{
		API:        api,
		Namespace:  *namespace,
		TaskQueue:  *taskQueue,
		Since:      time.Now().Add(-*since),
		Thresholds: thresh,
	}

	findings := runAll(ctx, runtimeall.Checks(), deps, stderr)
	sort.SliceStable(findings, func(i, j int) bool {
		if findings[i].Severity.Rank() != findings[j].Severity.Rank() {
			return findings[i].Severity.Rank() > findings[j].Severity.Rank()
		}
		if findings[i].Check != findings[j].Check {
			return findings[i].Check < findings[j].Check
		}
		return findings[i].Subject < findings[j].Subject
	})

	if err := emit(stdout, findings, *output); err != nil {
		fmt.Fprintf(stderr, "emit findings: %v\n", err)
		return 2
	}

	return exitCode(findings, *failOn)
}

// dial constructs a client.Client. API key path uses HTTPS / Cloud
// conventions; without a key we fall back to a plain insecure dial,
// which is what `temporal server start-dev` exposes.
func dial(address, namespace, apiKey string) (client.Client, error) {
	opts := client.Options{
		HostPort:  address,
		Namespace: namespace,
	}
	if apiKey != "" {
		opts.Credentials = client.NewAPIKeyStaticCredentials(apiKey)
		opts.HeadersProvider = headerInjector{"temporal-namespace": namespace}
	}
	return client.Dial(opts)
}

type headerInjector map[string]string

func (h headerInjector) GetHeaders(_ context.Context) (map[string]string, error) {
	out := make(map[string]string, len(h))
	for k, v := range h {
		out[k] = v
	}
	return out, nil
}

// runAll fans out every check; check-internal errors are logged to
// stderr but do not abort sibling checks.
func runAll(ctx context.Context, checks []runtime.Check, deps runtime.Deps, stderr io.Writer) []runtime.Finding {
	var (
		mu  sync.Mutex
		wg  sync.WaitGroup
		out []runtime.Finding
	)
	for _, c := range checks {
		c := c
		wg.Add(1)
		go func() {
			defer wg.Done()
			findings, err := c.Run(ctx, deps)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				fmt.Fprintf(stderr, "check %s: %v\n", c.Name(), err)
				return
			}
			out = append(out, findings...)
		}()
	}
	wg.Wait()
	return out
}

func emit(w io.Writer, findings []runtime.Finding, format string) error {
	switch strings.ToLower(format) {
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(struct {
			Findings []runtime.Finding `json:"findings"`
		}{findings})
	case "text", "":
		if len(findings) == 0 {
			_, err := fmt.Fprintln(w, "no findings")
			return err
		}
		for _, f := range findings {
			if _, err := fmt.Fprintf(w, "[%s] %s — %s: %s\n    %s\n", strings.ToUpper(string(f.Severity)), f.Check, f.Subject, f.Message, f.URL); err != nil {
				return err
			}
		}
		return nil
	}
	return errors.New("unsupported output format: " + format)
}

func exitCode(findings []runtime.Finding, failOn string) int {
	min := runtime.Severity(strings.ToLower(failOn))
	if min != runtime.SeverityFail && min != runtime.SeverityWarn && min != runtime.SeverityOK {
		min = runtime.SeverityFail
	}
	for _, f := range findings {
		if f.Severity.Rank() >= min.Rank() {
			return 1
		}
	}
	return 0
}
