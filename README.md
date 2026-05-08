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
    go run ./tools/temporallint/cmd/temporallint ./apps/...

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

## Layout

    tools/temporallint/
      cmd/temporallint/      multichecker entry point
      cmd/gen-rules/         generates RULES.md from all/ metadata
      plugin/                golangci-lint v2 module-plugin entry
      all/                   registry: Analyzers() + Registry()
      temporalctx/           shared helpers (workflow detection, rulekit)
      rules/<rule>/          per-rule directory
      RULES.md               generated coverage table
      go.mod                 sub-module — analyzer deps don't leak into apps/

## Why a sub-module?

Analyzer deps (`golang.org/x/tools`, `go.temporal.io/sdk/contrib/tools/workflowcheck`)
are heavy and not needed at runtime by `apps/`. The sub-module isolates
them so `go build ./apps/...` stays lean. The repo-root `go.work` joins
them back for unified `go test`.
