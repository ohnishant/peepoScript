# Testing

How the test layers are split and what each one is for.

## Layers

Unit tests cover the lexer, parser, evaluator, AST, runner path
resolution, and the REPL completer logic in isolation. They run fast on
any machine and need no terminal.

Integration tests (`cmd/repl/repl_integration_test.go`) drive the
compiled binary through a real pseudo-terminal: they type partial
keywords, press Tab, and assert on what the REPL renders back.

## Why the integration layer exists

The bugs that actually break autocomplete are invisible to unit tests:
readline losing its completer, a zero-width terminal silently disabling
completion, raw-mode key handling regressing. A completed keyword
appearing in the output is genuine proof, since the tests only ever
type a prefix; no unit test can fake that end-to-end path.

## Requirements and skip behavior

The pty layer needs a Linux kernel with `/dev/ptmx` (see
`cmd/repl/pty_linux_test.go`) and a Go toolchain on PATH to build the
binary. Everywhere else the tests skip themselves, so `go test ./...`
stays green off Linux. The GitHub CI runners satisfy both requirements,
so autocomplete regressions fail the build there.
