// Package cmd implements `temporallint cleanup-versions`. It drives
// the three phases (static discovery → runtime verification → optional
// apply) and emits a text or JSON report.
package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"go.temporal.io/sdk/client"

	"github.com/mattconzen/temporallint/cleanupversions"
	cuvruntime "github.com/mattconzen/temporallint/cleanupversions/runtime"
	"github.com/mattconzen/temporallint/cleanupversions/rewriter"
	"github.com/mattconzen/temporallint/cleanupversions/static"
)

// Main is the entry point. Returns the process exit code.
func Main(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("temporallint cleanup-versions", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var (
		address          = fs.String("address", "localhost:7233", "Temporal frontend gRPC address")
		namespace        = fs.String("namespace", "default", "Temporal namespace")
		apiKey           = fs.String("api-key", os.Getenv("TEMPORAL_API_KEY"), "Temporal API key")
		source           = fs.String("source", "./", "Go source root to scan")
		apply            = fs.Bool("apply", false, "rewrite source for Safe candidates (off by default)")
		output           = fs.String("output", "text", "report format: text | json")
		maxOpen          = fs.Int("max-open-workflows", 5000, "cap on open-workflow scan")
		unsafeAllowNoSrv = fs.Bool("unsafe-allow-no-server", false, "skip runtime verification (every candidate is Indeterminate; --apply is rejected)")
	)
	if err := fs.Parse(args); err != nil {
		return 2
	}

	candidates, err := static.Discover(*source)
	if err != nil {
		fmt.Fprintf(stderr, "discover: %v\n", err)
		return 2
	}

	var reports []cleanupversions.SafetyReport
	if *unsafeAllowNoSrv {
		if *apply {
			fmt.Fprintln(stderr, "--apply requires server verification; refusing to run with --unsafe-allow-no-server")
			return 2
		}
		for _, c := range candidates {
			d := cleanupversions.DecisionIndeterminate
			detail := "server verification skipped"
			if c.Reason != "" {
				d = cleanupversions.DecisionSkip
				detail = c.Reason
			}
			reports = append(reports, cleanupversions.SafetyReport{Candidate: c, Decision: d, Detail: detail})
		}
	} else {
		c, err := dial(*address, *namespace, *apiKey)
		if err != nil {
			fmt.Fprintf(stderr, "dial temporal: %v\n", err)
			return 2
		}
		defer c.Close()
		api := newAdapter(c, *namespace)
		v := &cuvruntime.Verifier{API: api, Namespace: *namespace, MaxOpen: *maxOpen}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		reports, err = v.Verify(ctx, candidates)
		if err != nil {
			fmt.Fprintf(stderr, "verify: %v\n", err)
			return 2
		}
	}

	applied := false
	if *apply {
		// Refuse the apply if any candidate is not Safe.
		for _, r := range reports {
			if r.Decision != cleanupversions.DecisionSafe {
				fmt.Fprintf(stderr, "--apply refused: at least one candidate is %s (%s); resolve before retrying\n", r.Decision, r.Candidate.ChangeID)
				if err := emit(stdout, reports, false, *output); err != nil {
					fmt.Fprintln(stderr, err)
				}
				return 1
			}
		}
		touched, err := rewriter.Apply(reports)
		if err != nil {
			fmt.Fprintf(stderr, "apply: %v\n", err)
			return 2
		}
		applied = len(touched) > 0
		fmt.Fprintf(stderr, "applied edits to %d file(s): %v\n", len(touched), touched)
	}

	if err := emit(stdout, reports, applied, *output); err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	for _, r := range reports {
		if r.Decision == cleanupversions.DecisionUnsafe {
			return 1
		}
	}
	return 0
}

func emit(w io.Writer, reports []cleanupversions.SafetyReport, applied bool, format string) error {
	switch strings.ToLower(format) {
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(cleanupversions.Report{Reports: reports, Applied: applied})
	case "text", "":
		if len(reports) == 0 {
			_, err := fmt.Fprintln(w, "no GetVersion candidates found")
			return err
		}
		for _, r := range reports {
			_, err := fmt.Fprintf(w, "[%s] change_id=%q file=%s:%d  %s\n",
				strings.ToUpper(string(r.Decision)),
				r.Candidate.ChangeID,
				r.Candidate.Pos.Filename,
				r.Candidate.Pos.Line,
				r.Detail,
			)
			if err != nil {
				return err
			}
		}
		if applied {
			fmt.Fprintln(w, "applied: source files rewritten")
		}
		return nil
	}
	return errors.New("unsupported output format: " + format)
}

func dial(address, namespace, apiKey string) (client.Client, error) {
	opts := client.Options{HostPort: address, Namespace: namespace}
	if apiKey != "" {
		opts.Credentials = client.NewAPIKeyStaticCredentials(apiKey)
	}
	return client.Dial(opts)
}
