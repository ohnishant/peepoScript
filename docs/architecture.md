# Architecture

How peepoScript is organized and where changes land. Written for agents
working in this repo; humans can read it too.

## Pipeline

Source text flows through four stages. Each stage has its own package,
and each only talks to the one below it.

```
source -> token -> lexer -> parser -> ast -> evaluator -> output
```

- `cmd/token` defines the token types and maps keyword literals
  (`PepoG`, `Wokege`, ...) to token types.
- `cmd/lexer` turns source into tokens. Keywords are resolved through
  `token.LookupIdent`; anything else becomes an IDENT.
- `cmd/parser` turns tokens into an AST. It owns error recovery for
  statement parsing and collects all parse errors per run instead of
  stopping at the first one.
- `cmd/ast` defines the node types. Expressions live separately in
  `expressions.go`.
- `cmd/evaluator` walks the AST. Everything it produces is an
  `Object` (see `object.go`). Bindings live in `Environment`
  (`environment.go`), which chains scopes for function calls.

## Entry points

Three frontends share one execution path through `cmd/runner`:

- **REPL** (`cmd/repl`) reads lines interactively, supports tab
  completion over keywords plus builtins, and echoes each expression's
  value. The pty plumbing used by integration tests lives beside it in
  `pty_linux_test.go`.
- **CLI** (`main.go` at the repo root) handles `-c` inline source and
  script file arguments. It resolves script paths via
  `runner.ResolvePath`, which appends `.peepo` when the bare path does
  not exist.
- **WASM** (`cmd/wasm`, build tag `js && wasm`) exposes two globals to
  JavaScript: `peepoRun(code)` returning `{ok, output, value, errors}`,
  and `peepoKeywords()` listing keywords plus builtins for the page's
  completion and emote rendering. The environment persists across runs,
  so bindings survive between calls.

`cmd/runner.Run(code, env)` is the embedder-facing API all three use.
It returns typed errors: `*ParseError` (collected diagnostics) and
`*RuntimeError` (wrapping an evaluator `Error` object). Frontends match
on these types to render errors differently.

## Output injection

`peepoChat` never writes to hardcoded stdout. Output goes through an
`io.Writer` stored on the evaluator's builtin set, set at environment
creation (`evaluator.NewEnvironmentWithOut`). The CLI passes stdout,
the REPL passes its own output writer, and the WASM build captures into
a string builder returned per call. Any new builtin that prints must go
through that writer, not `os.Stdout`.

## Adding a language feature

A new keyword touches, in order: `cmd/token/token.go` (literal to token
type), `cmd/lexer` (if a new token shape is needed), `cmd/parser`
(precedence and parse rule), `cmd/ast` (node), `cmd/evaluator` (case in
`Eval`). A new builtin only touches `cmd/evaluator/builtins.go`.

Naming note: new keywords, error messages, and builtins need emote-name
approval from the repo owner before they are written into code. See
AGENTS.md; established emotes already in the language need no approval
when reused.

## Tooling around the language

- `cmd/emotefetch` vendors keyword emotes from 7tv into `web/emotes`
  so the playground renders offline. Emote ids resolve in priority
  order: hand-pinned ids, towdan's channel set, then global search.
- `scripts/build-web.sh` builds the WASM bundle and copies static
  assets into `web/`. Pushes to `main` deploy the playground to GitHub
  Pages (`.github/workflows/pages.yml`).
