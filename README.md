# peepoScript

A scripting language made for ![peepo](assets/peepo.png) by ![peepo](assets/peepo.png).

## Contents

- [Running](#running)
  - [REPL](#repl)
  - [Scripts](#scripts)
  - [Browser playground](#browser-playground)
- [The language](#the-language)
  - [Comments](#comments)
  - [Bindings](#bindings)
  - [Operators](#operators)
  - [Booleans and conditionals](#booleans-and-conditionals)
  - [Loops](#loops)
  - [Functions](#functions)
  - [Lists](#lists)
  - [Builtins](#builtins)
- [Examples](#examples)
- [Project layout](#project-layout)
- [Tests](#tests)

## Running

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

In an interactive terminal, Tab completes language keywords and builtins. Type `peepoCo` and hit Tab to get `peepoCookie`.

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

### Browser playground

The interpreter also compiles to WebAssembly and runs on a static page. Editor on the left with tab completion, output panel on the right, and keywords plus builtins vendored as 7tv emotes so they render while you type.

Build the wasm bundle and serve the page locally:

```sh
./scripts/build-web.sh
cd web && python3 -m http.server 8080
```

Pushes to `main` deploy the playground to GitHub Pages automatically (`.github/workflows/pages.yml`).

### Examples

The `examples/` directory has runnable scripts, ordered simple to less simple. Each one runs as-is:

```sh
./peepo examples/arithmetic.peepo
```

- `arithmetic.peepo`: bindings, operator-first arithmetic, truncating division
- `strings.peepo`: concatenation and `peepoMeasure`
- `conditionals.peepo`: `Hmmge`/`peepoShrug`, including a nested branch
- `countdown.peepo`: recursion in place of a loop
- `fibonacci.peepo`: two functions, one computing and one driving the recursion

## The language

Expressions are written operator-first: the operator comes before its operands. `peepoFriendship 1 2` means `1 + 2`, and calling a function is just writing its name followed by arguments. That is also why blocks don't need parentheses around conditions.

### Comments

`#` starts a comment; everything up to the end of the line is ignored.

```
PepoG x 10.  # declare
# this whole line is a comment
```

### Bindings

```
PepoG answer 42.          # declare
peepoCookie answer 43.    # assign
```

### Operators

| Keyword            | Meaning      |
| ------------------ | ------------ |
| `peepoFriendship`  | addition     |
| `PepegaCredit`     | subtraction  |
| `mitosis`          | multiplication |
| `peepoBye`         | division     |
| `peepoLessThan`    | less than    |
| `peepoGreaterThan` | greater than |
| `Scoots`           | equality     |

Operators nest freely, so `mitosis peepoFriendship 1 2 3` is `(1 + 2) * 3`.

### Booleans and conditionals

`NODDERS` is true, `NOPERS` is false. Conditions have to evaluate to a boolean; there is no implicit truthiness.

```
Hmmge Scoots x 10 Wokege
    peepoChat "ten".
Bedge peepoShrug Wokege
    peepoChat "not ten".
Bedge
```

### Loops

`peepoJuice` runs a `Wokege`/`Bedge` block while its condition holds. The init and step clauses are optional `PepoG` or `peepoCookie` statements; leave both out and the loop is a plain while. Assignments reach the scope a variable was declared in, so the body can accumulate into outer bindings.

```
peepoJuice PepoG i 0. peepoLessThan i 3. peepoCookie i peepoFriendship i 1. Wokege
    peepoChat i.
Bedge

PepoG n 3.
peepoJuice peepoGreaterThan n 0. peepoCookie n PepegaCredit n 1. Wokege
    peepoChat n.
Bedge
```

### Functions

Functions are values. The body evaluates to whatever its last expression produces, so there is no return keyword.

```
PepoG add SadgeBusiness a b Wokege
    peepoFriendship a b.
Bedge

PepoG result add 40 2.   # 6
```

### Lists

`Thinking1` opens a list and `Thinking2` closes it, for both literals and indexing.

```
PepoG xs Thinking1 3, 1, 4 Thinking2.
peepoChat xs.                          # [3, 1, 4]
peepoChat xs Thinking1 0 Thinking2.    # 3
peepoChat peepoMeasure xs.             # 3
```

Lists nest freely; chain the brackets to dig into them: `grid Thinking1 1 Thinking2 Thinking1 0 Thinking2`. Indexing out of range or with a non-integer is a runtime error. Combine with `peepoJuice` to walk a list: see [examples/lists.peepo](examples/lists.peepo).

### Builtins

- `peepoChat arg1 arg2 ...` prints its arguments separated by spaces.
- `peepoMeasure str` returns the length of a string or a list.

## Project layout

Go modules live under `cmd/`, one directory per stage of the pipeline plus three frontends (REPL, CLI, WASM). See [docs/architecture.md](docs/architecture.md) for the map and where changes land when adding language features.

## Tests

```sh
go test ./...
```

Unit tests cover each module in isolation; integration tests drive the compiled binary through a real pseudo-terminal to prove autocomplete works end to end. They skip themselves where `/dev/ptmx` isn't available, so the suite stays green on any machine. Details in [docs/testing.md](docs/testing.md).
