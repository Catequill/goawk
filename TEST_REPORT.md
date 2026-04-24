# Test Report

Environment:
- Go: `go1.26.2 darwin/arm64`
- Command: `go test ./...`
- Date: 2026-04-24

## Package results

| Package | Result | Time |
| --- | --- | --- |
| `github.com/benhoyt/goawk` | FAIL | 2.537s |
| `github.com/benhoyt/goawk/internal/ast` | ok | 0.167s |
| `github.com/benhoyt/goawk/internal/compiler` | ok | 0.471s |
| `github.com/benhoyt/goawk/internal/cover` | ok | 0.627s |
| `github.com/benhoyt/goawk/internal/parseutil` | ok | 0.779s |
| `github.com/benhoyt/goawk/internal/resolver` | ok | 0.937s |
| `github.com/benhoyt/goawk/interp` | FAIL | 1.381s |
| `github.com/benhoyt/goawk/lexer` | ok | 1.302s |
| `github.com/benhoyt/goawk/parser` | ok | 1.208s |

7/9 packages pass. The two failing packages contain the "compare goawk output against a reference AWK" tests.

## Root-cause of failures

Both failing packages are blocked by a missing `gawk` executable on this host:

```
goawk_test.go:108: error running gawk: exec: "gawk": executable file not found in $PATH
```

Failing tests in `github.com/benhoyt/goawk`:
- `TestAWK` — runs every script in `testdata/` through `gawk` to build the expected output.
- `TestCommandLine` — subtests invoke `gawk` as the reference implementation.
- `TestDevStdout` — compares against `gawk`.
- `TestFILENAME` — compares against `gawk`.

Failing test in `github.com/benhoyt/goawk/interp`:
- `TestInterp` — same gawk-comparison pattern.

The `-awk` flag (default `gawk`) selects the reference binary. Retrying with `-awk=/usr/bin/awk` (macOS BSD awk) narrows failures but surfaces real divergences (BSD awk rejects `-E`, `-w`; `TestFILENAME` behaves differently) — BSD awk is not a valid substitute, so these are not goawk regressions.

## Reproduction

```bash
go test ./...
# to run the gawk-dependent suites, install gawk first:
#   brew install gawk
# then re-run the same command.
```

## Conclusion

All pure-Go packages (lexer, parser, resolver, compiler, ast, cover, parseutil) pass. The root and `interp` packages fail solely because the host lacks `gawk`; installing gawk is expected to restore a green run.
