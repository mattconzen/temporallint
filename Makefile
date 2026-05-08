# temporallint helper targets.
#
# Run from the repo root via:
#   make -C tools/temporallint <target>
#
# Or with a fully-qualified Makefile path:
#   make -f tools/temporallint/Makefile <target>

.PHONY: test verify-tdd-red gen-rules canary all

# Default: run the full unit-test suite.
test:
	go test ./tools/temporallint/...

# Audit the inverse-red property: for every rule under rules/<name>/,
# disable the rule's logic and confirm the rule's test fails. Exit code
# 1 if any rule's test passes with the rule disabled (the test isn't
# actually exercising the rule).
#
# Coverage: static analyzers under rules/<name>/ only. Runtime checks
# (runtime/checks/<name>) and cleanup-versions phases use different
# test patterns and are NOT covered by this target — their inverse-red
# guarantees come from explicit assertions in the test bodies.
verify-tdd-red:
	go run ./tools/temporallint/cmd/verify-tdd-red

# Regenerate RULES.md from the registry and fail if it changes (CI use).
gen-rules:
	go run ./tools/temporallint/cmd/gen-rules > tools/temporallint/RULES.md
	@git diff --exit-code -- tools/temporallint/RULES.md || \
		(echo "RULES.md is stale — commit the regenerated file"; exit 1)

# Canary: run the static analyzer over apps/. continue-on-error per
# the existing CI policy.
canary:
	-go run ./tools/temporallint/cmd/temporallint ./apps/...

all: test verify-tdd-red gen-rules
