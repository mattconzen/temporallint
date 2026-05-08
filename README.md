# temporallint

Static analysis for Temporal anti-patterns, modeled on the catalogue in
[jlegrone/100-temporal-mistakes](https://github.com/jlegrone/100-temporal-mistakes).

## What it does

`temporallint` is a Go static-analysis suite. Each anti-pattern lives in
its own analyzer under `rules/<rule>/`. The list of every documented
mistake (implemented or not) is in [`RULES.md`](./RULES.md), generated
from `all/all.go` + `all/planned.go`.

This module covers **Batch 1 + Batch 2** of the implementation plan:
the foundation, plus 12 hand-rolled non-determinism rules layered on top
of Temporal's own [`workflowcheck`](https://pkg.go.dev/go.temporal.io/sdk/contrib/tools/workflowcheck/workflow)
SSA call-graph checker. The remaining batches (timeouts, activity shape,
cancellation, design, ops) are tracked as `Planned` entries in `RULES.md`.

## Running locally

    # From the repo root (go.work joins this sub-module)
    go test ./tools/temporallint/...

    # Static analysis against Go packages
    go run ./tools/temporallint/cmd/temporallint ./apps/...

    # Runtime checks against a live Temporal server
    go run ./tools/temporallint/cmd/temporallint runtime \
        --address localhost:7233 --namespace default --since 24h

The same binary handles both modes via argv dispatch: `runtime` as the
first positional arg routes into the runtime subcommand, anything else
(including `./...`) falls through to the static-analysis multichecker.

## Adding a rule

1. `mkdir -p rules/<name>/testdata/src/{violation,clean}`
2. Copy the workflow stub:
   `cp temporalctx/testdata_workflow_stub.go.tmpl rules/<name>/testdata/src/go.temporal.io/sdk/workflow/workflow.go`
   (drop the leading template comment so it's a valid Go file)
3. Write `rule.go` exporting `var Analyzer = &analysis.Analyzer{...}`. For
   "ban this qualified call inside workflow code" rules, use
   `temporalctx.RunCallBans` — it already handles the workflow-detection
   walk and the type-info / fallback matching.
4. Write `rule_test.go` calling `analysistest.Run(t, dir, Analyzer, "violation")`
   then `... "clean"`.
5. Write the violation fixture with `// want "regexp"` markers. This is
   the TDD-red assertion: deleting the analyzer's body must make the test
   fail.
6. Write the clean fixture — must produce zero diagnostics.
7. Register the rule in `all/all.go` (the `implemented` list) with a
   `MistakeURL` pointing at the corresponding section of the source repo.
8. Regenerate the docs: `go run ./tools/temporallint/cmd/gen-rules > tools/temporallint/RULES.md`

## golangci-lint integration

`plugin/plugin.go` exposes the analyzer registry via the v2 module-plugin
entry point. Wire it up with a `.custom-gcl.yml` at the repo root:

    version: v2.0.0
    plugins:
      - module: github.com/mattconzen/monorepo/tools/temporallint
        path: ./tools/temporallint

Then `golangci-lint custom` builds a `custom-gcl` binary that includes
every temporallint rule alongside golangci-lint's built-ins.

## Runtime subcommand

`temporallint runtime` connects to a Temporal server and verifies the
mistakes that static analysis can't see — history size, event count,
individual payload size, missing workflow timeouts. It paginates
`ListWorkflowExecutions` over the configured time window and calls
`DescribeWorkflowExecution` / `GetWorkflowHistory` per workflow.

Flags:

| Flag | Default | Notes |
|------|---------|-------|
| `--address` | `localhost:7233` | Temporal frontend gRPC address |
| `--namespace` | `default` | Temporal namespace |
| `--api-key` | `$TEMPORAL_API_KEY` | API key for Cloud or any gateway with API-key support |
| `--task-queue` | (none) | Optional filter |
| `--since` | `24h` | Look-back window |
| `--threshold-config` | (none) | YAML override path; see `runtime/thresholds/default.yaml` |
| `--output` | `text` | `text` or `json` |
| `--fail-on` | `fail` | Minimum severity that triggers non-zero exit (`ok` / `warn` / `fail`) |

Each finding includes the check name, severity, subject (workflowID/runID
or event identifier), human message, and a link back to the corresponding
section of [100-temporal-mistakes](https://github.com/jlegrone/100-temporal-mistakes).

Tests use a hand-rolled `runtime/fakeapi` test double rather than
mockgen — every check has a TDD-inverted pair:
`TestViolation` (synthetic API returns above-threshold response → expect
`Severity=fail` Finding), `TestClean` (response within threshold → expect
no Finding). Mirror of the static analyzers' `testdata/src/violation` /
`testdata/src/clean` pattern.

## Layout

    tools/temporallint/
      cmd/temporallint/      argv dispatcher (static multichecker + runtime)
      cmd/gen-rules/         generates RULES.md from all/ metadata
      plugin/                golangci-lint v2 module-plugin entry
      all/                   registry: Analyzers() + Registry()
      temporalctx/           shared helpers (workflow detection, rulekit)
      rules/<rule>/          per-rule static-analysis directory
      runtime/
        cmd/                 runtime subcommand impl
        all/                 registry of runtime.Check
        checks/<name>/       per-check directory
        thresholds/          defaults + YAML loader
        internal/fakeapi/    test double for WorkflowAPI
        internal/temporaladapter/   client.Client → WorkflowAPI bridge
      RULES.md               generated coverage table
      go.mod                 sub-module — analyzer + Temporal deps don't leak into apps/

## Why a sub-module?

Analyzer deps (`golang.org/x/tools`, `go.temporal.io/sdk/contrib/tools/workflowcheck`)
are heavy and not needed at runtime by `apps/`. The sub-module isolates
them so `go build ./apps/...` stays lean. The repo-root `go.work` joins
them back for unified `go test`.
