# Modern Go Features by Version

Use this file to choose version-appropriate syntax and stdlib APIs.

## Rules

- Detect and honor the module Go version before coding.
- Use features only up to and including the target version.
- Prefer clearer standard library helpers over manual patterns.
- If uncertain about feature availability, check the release notes for that version.

## Go 1.13+

- Use `errors.Is` and `errors.As` for wrapped error matching.

## Go 1.18+

- Use `any` instead of `interface{}` where appropriate.
- Use generics where they reduce duplication without obscuring intent.
- Use `strings.Cut` and `bytes.Cut` for delimiter splitting.

## Go 1.19+

- Use typed atomics (`atomic.Bool`, `atomic.Int64`, `atomic.Pointer[T]`) when atomic coordination is required.

## Go 1.20+

- Use `errors.Join` for combining independent errors.
- Use `context.WithCancelCause` and `context.Cause` when cause propagation is relevant.
- Use `strings.CutPrefix` and `strings.CutSuffix` for explicit prefix/suffix parsing.
- Use `strings.Clone` and `bytes.Clone` when ownership isolation is needed.

## Go 1.21+

- Use `min`, `max`, and `clear` built-ins when they improve clarity.
- Use `slices` and `maps` helper packages for common collection operations.
- Use `sync.OnceFunc` and `sync.OnceValue` for one-time initialization helpers.
- Use `log/slog` for structured logging where it fits the project conventions.

## Go 1.22+

- Use `for i := range n` when integer range loops improve readability.
- Rely on fixed loop variable capture behavior in closures.
- Use enhanced `net/http` `ServeMux` patterns and `r.PathValue` when available.

## Go 1.23+

- Use iterator-capable collection helpers where they simplify code.

## For Newer Versions

- Review `https://go.dev/doc/devel/release` for each target version.
- Add newer APIs only when confirmed available in the module target version.

## Fallback Guidance

- If a helper is unavailable, use the clearest compatible idiom.
- Prefer portability and readability over aggressive feature usage.
- Mention version constraints only when the choice is non-obvious.
