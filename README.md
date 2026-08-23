# peepoScript

A scripting language made for ![peepo](assets/peepo.png) by ![peepo](assets/peepo.png).

## Running the interpreter

You need [Go](https://go.dev) 1.22 or newer.

```sh
go run .
```

Or build a binary first:

```sh
go build -o peepo .
./peepo
```

### REPL

With no arguments the binary starts the REPL:

```
Peepo Script - 0.1.0 | REPL

>> PepoG x 10.
>> peepoFriendship x 5.
15
```

Statements end with a fullstop. The result of the last expression gets printed back at you, which doubles as your print statement when you just want to check something.

In an interactive terminal, Tab completes language keywords. Type `peepoC` and hit Tab to get `peepoCookie`:

Blocks open with `Wokege` and close with `Bedge`. If you hit enter before closing one, the REPL drops to a `...` prompt and keeps reading until the block is done:

```
>> SadgeBusiness greet name Wokege
...   peepoChat "hello" name.
... Bedge
>> greet "peepo"
hello peepo
```

### Scripts

The `-c` flag runs source straight from the command line:

```sh
./peepo -c 'PepoG x 10. peepoChat x.'
10
```

Give it a path instead and it runs that file. A path without an extension gets `.peepo` appended when the bare path doesn't exist, so both of these work:

```sh
./peepo scripts/greeting.peepo
./peepo scripts/greeting      # finds greeting.peepo
```

Script mode does not echo the value of the last expression back the way the REPL does. If a script should say something, it says it through `peepoChat`.

Parse errors, runtime errors, and missing files exit with status 1 and print on stderr. A clean run exits 0.

### What's next

The plan is to compile the interpreter to WASM so scripts can run on a webpage (see issue #13 for details). That work builds on two choices already in place: script execution lives in `cmd/runner`, which has no terminal dependencies, and `peepoChat` writes to an injected writer instead of hardcoded stdout, so a browser build can route output into the page.

## The language

Expressions are written operator-first: the operator comes before its operands. `peepoFriendship 1 2` means `1 + 2`, and calling a function is just writing its name followed by arguments. That is also why blocks don't need parentheses around conditions.

### Bindings

```
PepoG answer 42.          # declare
peepoCookie answer 43.    # assign
```

### Operators

| Keyword           | Meaning         |
| ----------------- | --------------- |
| `peepoFriendship` | addition        |
| `PepegaCredit`    | subtraction     |
| `mitosis`         | multiplication  |
| `peepoBye`        | division        |
| `peepoLessThan`   | less than       |
| `peepoGreaterThan`| greater than    |
| `Scoots`          | equality        |
| `peepoJuice`      | not             |

Operators nest freely, so `mitosis peepoFriendship 1 2 3` is `(1 + 2) * 3`.

### Booleans

`NODDERS` is true, `NOPERS` is false.

### Conditionals

The condition has to evaluate to a boolean. There's no implicit truthiness.

```
Hmmge Scoots x 10 Wokege
    peepoChat "ten".
Bedge peepoShrug Wokege
    peepoChat "not ten".
Bedge
```

### Functions

Functions are values. The body evaluates to whatever its last expression produces, so there is no return keyword.

```
PepoG add SadgeBusiness a b Wokege
    peepoFriendship a b.
Bedge

PepoG result add add 1 2 3.   # 6
```

### Builtins

- `peepoChat arg1 arg2 ...` prints its arguments separated by spaces.
- `peepoMeasure str` returns the length of a string.

## Tests

```sh
go test ./...
```

One command runs everything, in two layers:

**Unit tests** cover the lexer, parser, evaluator, and the completer logic in isolation. Fast, no terminal needed.

**Integration tests** (`cmd/repl/repl_integration_test.go`) drive the compiled binary through a real pseudo-terminal: they type partial keywords, press Tab, and assert on what the REPL renders back. These exist because the bugs that actually break autocomplete — readline losing its completer, a zero-width terminal silently disabling completion, raw-mode key handling regressing — are invisible to unit tests. A completed keyword appearing in the output is genuine proof, since the tests only ever type a prefix.

Requirements for the integration layer: a Linux kernel with `/dev/ptmx` (the pty plumbing lives in `cmd/repl/pty_linux_test.go`) and a Go toolchain on PATH to build the binary. Everywhere else those tests skip themselves, so `go test ./...` stays green on any machine. GitHub's Ubuntu CI runners satisfy both requirements out of the box — no extra workflow config is needed; if autocomplete ever breaks for real users, these tests fail the build.
