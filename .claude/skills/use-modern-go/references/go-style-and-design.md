# Go Style and Design Reference

## Table of Contents

1. [Core Priorities](#core-priorities)
2. [Formatting and File Hygiene](#formatting-and-file-hygiene)
3. [Naming Rules](#naming-rules)
4. [Package and Module Design](#package-and-module-design)
5. [API Design and Compatibility](#api-design-and-compatibility)
6. [Type and Method Design](#type-and-method-design)
7. [Interfaces and Abstractions](#interfaces-and-abstractions)
8. [Error Design and Handling](#error-design-and-handling)
9. [Context Usage](#context-usage)
10. [Observability and Operability](#observability-and-operability)
11. [Security and Input Boundaries](#security-and-input-boundaries)
12. [Data Semantics: Nil, Empty, Copying](#data-semantics-nil-empty-copying)
13. [Comments and Documentation](#comments-and-documentation)
14. [Dependency and Scalability Guidance](#dependency-and-scalability-guidance)
15. [Performance Posture](#performance-posture)
16. [Review Checklist](#review-checklist)

## Core Priorities

Apply this order when making tradeoffs:

1. Clarity
2. Simplicity
3. Concision
4. Maintainability
5. Consistency

Prefer boring, explicit code over clever code.

## Formatting and File Hygiene

- Run `gofmt` for every edit.
- Keep imports grouped and deterministic.
- Keep signatures and declarations easy to scan.
- Avoid indentation patterns that hide error paths.

## Naming Rules

- Use lower-case package names without underscores.
- Keep package names short, descriptive, and specific.
- Avoid `util`, `common`, `helper`, and similar catch-all names.
- Use short receiver names (usually 1-2 letters), consistent per type.
- Avoid `Get`/`Set` prefixes unless domain language requires them.
- Keep initialism casing consistent (`ID`, `URL`, `HTTP`).
- Keep variable names proportional to scope.

## Package and Module Design

- Organize packages around domain behavior, not only technical layers.
- Keep package public surfaces minimal and stable.
- Hide implementation details with unexported symbols.
- Use `internal/` to enforce ownership boundaries when helpful.
- Keep dependency direction explicit and acyclic.
- Prefer one module per deployable boundary unless there is a strong reason to split.

## API Design and Compatibility

- Make zero values useful when possible.
- Make invalid states hard to represent.
- Favor additive evolution over breaking changes.
- Document any intentional breaking change and migration path.
- Keep side effects explicit in names and docs.
- Use option structs when argument lists grow.
- Return concrete types from constructors unless abstraction is required.

## Type and Method Design

- Use value receivers for small immutable-like types.
- Use pointer receivers for mutation, large structs, or method-set consistency.
- Keep receiver style consistent per type.
- Avoid exporting mutable struct fields by default.
- Prefer concrete types over `map[string]any` for stable schemas.
- Use generics when they reduce duplication and remain readable.

## Interfaces and Abstractions

- Define interfaces where consumed.
- Keep interfaces behavior-oriented and small.
- Avoid speculative abstraction before multiple implementations exist.
- Prefer concrete parameters until substitution is needed.
- Explain non-obvious abstraction boundaries.

## Error Design and Handling

- Return `error` as the last return value.
- Handle errors immediately; prefer early returns.
- Wrap with `%w` and actionable context.
- Use `errors.Is`/`errors.As` for semantic checks.
- Keep error strings lowercase and without punctuation.
- Avoid duplicate logging and returning of the same error at the same boundary.

## Context Usage

- Accept `context.Context` as first parameter for request-scoped operations.
- Propagate context through call chains and goroutines.
- Respect cancellation/deadlines in all blocking paths.
- Do not store context in structs.
- Use `context.Background()` only at true process entrypoints.

## Observability and Operability

- Emit structured logs at boundaries (I/O, RPC, retries, failures).
- Keep logs sparse in hot paths; prefer metrics/tracing for high-volume signals.
- Include stable identifiers (`request_id`, `user_id`, `job_id`) when available.
- Avoid logging secrets, tokens, and raw sensitive payloads.
- Make timeouts, retries, and backoff explicit and configurable.

## Security and Input Boundaries

- Validate and normalize external input at boundaries.
- Fail closed on malformed or unauthorized inputs.
- Prefer allowlists over blocklists where practical.
- Keep authn/authz decisions close to request boundaries.
- Use least-privilege defaults for credentials and dependency scope.

## Data Semantics: Nil, Empty, Copying

- Be explicit about nil vs empty semantics.
- Return nil slices/maps for absence unless API contract requires empty.
- Clone/copy when ownership changes across package or goroutine boundaries.
- Document mutability and ownership expectations.

## Comments and Documentation

- Add doc comments for exported symbols.
- Start doc comments with the symbol name.
- Explain invariants, constraints, and side effects.
- Document cleanup responsibilities (`Close`, `Stop`, `Cancel`).
- Document error contracts for sentinel or typed errors.

## Dependency and Scalability Guidance

- Prefer stdlib first; add third-party dependencies intentionally.
- Justify each dependency with clear value and maintenance cost.
- Minimize transitive dependency surface area.
- Pin and review module updates with compatibility and security in mind.
- Run `go mod tidy` when dependency or import surfaces change.
- Design APIs and package boundaries for additive growth.

## Performance Posture

- Optimize after measurement, not intuition.
- Use benchmarks for hot paths and compare before/after.
- Use CPU/memory/block/mutex profiles before major refactors.
- Avoid micro-optimizations that reduce readability without measurable gain.

## Review Checklist

- Is naming clear and idiomatic?
- Is the package/API surface minimal and cohesive?
- Are compatibility and migration risks addressed?
- Is error handling actionable and chain-preserving?
- Is context propagated correctly?
- Are observability and timeout/retry decisions explicit?
- Are dependency choices justified?
- Is complexity supported by measurement?
