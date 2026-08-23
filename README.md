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

There is no file mode yet, so everything happens in the REPL:

```
Peepo Script - 0.1.0 | REPL

>> PepoG x 10.
>> peepoFriendship x 5.
15
```

Statements end with a fullstop. The result of the last expression gets printed back at you, which doubles as your print statement when you just want to check something.

Blocks open with `Wokege` and close with `Bedge`. If you hit enter before closing one, the REPL drops to a `...` prompt and keeps reading until the block is done:

```
>> SadgeBusiness greet name Wokege
...   peepoChat "hello" name.
... Bedge
>> greet "peepo"
hello peepo
```

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
