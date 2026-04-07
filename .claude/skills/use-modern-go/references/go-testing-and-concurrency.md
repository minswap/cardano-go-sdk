# Go Testing and Concurrency Reference

## Table of Contents

1. [Validation Matrix](#validation-matrix)
2. [Testing Strategy](#testing-strategy)
3. [Unit Test Structure](#unit-test-structure)
4. [Table-Driven Tests and Subtests](#table-driven-tests-and-subtests)
5. [Assertions and Comparisons](#assertions-and-comparisons)
6. [Integration and End-to-End Tests](#integration-and-end-to-end-tests)
7. [Fuzzing and Property Coverage](#fuzzing-and-property-coverage)
8. [Benchmarking and Profiling](#benchmarking-and-profiling)
9. [Concurrency Ownership Rules](#concurrency-ownership-rules)
10. [Context, Cancellation, and Timeouts](#context-cancellation-and-timeouts)
11. [Resource Cleanup and Leak Prevention](#resource-cleanup-and-leak-prevention)
12. [Reliability Checklist](#reliability-checklist)

## Validation Matrix

Use the smallest set that proves correctness for the change:

```bash
go test ./...
go vet ./...
go test -race ./...                  # concurrency touched
go test -fuzz=Fuzz -run=^$ ./...     # parser/decoder/input-heavy changes
go test -bench . ./...               # performance-sensitive changes
govulncheck ./...                    # security and dependency checks
```

## Testing Strategy

- Prefer fast deterministic unit tests.
- Add integration tests at transport and storage boundaries.
- Keep end-to-end tests for critical business flows.
- Make failures actionable with inputs and expected/actual context.

## Unit Test Structure

- Name tests by behavior (`TestParseRejectsInvalidHeader`).
- Follow arrange-act-assert flow.
- Use `t.Helper()` in shared helpers.
- Keep one failure reason per assertion block when practical.
- Avoid shared mutable globals across tests.

## Table-Driven Tests and Subtests

- Use table-driven tests for repeated logic patterns.
- Use stable concise subtest names.
- Capture loop variables correctly in closures/goroutines.
- Include key values in failure messages, not only subtest names.

## Assertions and Comparisons

- Use plain Go comparisons for simple checks.
- Use structural diffs (`cmp.Diff` or equivalent) for complex values.
- Compare errors with `errors.Is`/`errors.As`.
- Avoid brittle string matching on wrapped errors unless required.

## Integration and End-to-End Tests

- Prefer real transports with test servers when feasible.
- Verify serialization, retries, deadlines, and idempotency boundaries.
- Keep test fixtures isolated and reproducible.
- Use `testdata/` for stable fixture files.

## Fuzzing and Property Coverage

- Use fuzzing for parser-like code and untrusted input boundaries.
- Seed corpus with known edge and regression cases.
- Promote crashing/failing inputs into deterministic regression tests.
- Define invariants explicitly for property-style expectations.

## Benchmarking and Profiling

- Benchmark only hot paths and allocation-sensitive code.
- Use `b.ReportAllocs()` when allocation behavior matters.
- Prevent dead-code elimination by consuming outputs.
- Profile before major performance refactors.
- Compare benchmark baselines before accepting optimization complexity.

## Concurrency Ownership Rules

- Make goroutine owner and shutdown condition explicit.
- Use bounded worker pools/backpressure for untrusted or bursty workloads.
- Prefer `errgroup`/`WaitGroup` patterns to avoid orphan goroutines.
- Guard shared mutable state with mutexes/atomics and document invariants.
- Avoid copying structs that contain mutexes.
- Avoid fire-and-forget goroutines in library code.

## Context, Cancellation, and Timeouts

- Thread context through concurrent work and I/O boundaries.
- Respect context cancellation in loops and blocking operations.
- Ensure child goroutines stop when parent context is done.
- Make timeout intent explicit in both code and tests.
- Use cause-aware cancellation APIs when supported and useful.

## Resource Cleanup and Leak Prevention

- Close resources with `defer` near acquisition.
- Use `t.Cleanup` for test resource lifecycle.
- Document caller cleanup responsibilities in public APIs.
- Avoid `os.Exit` in code under test.
- Stop tickers/timers when lifecycle requires explicit cleanup.

## Reliability Checklist

- Do tests fail deterministically with useful diagnostics?
- Do race checks pass when concurrency is involved?
- Are goroutine lifecycles explicit and bounded?
- Are timeout/cancellation behaviors covered in tests?
- Are security-sensitive input paths fuzzed where relevant?
- Are vulnerability checks run for changed dependencies/surfaces?
