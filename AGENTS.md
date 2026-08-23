# AGENTS.md

## Language design

peepoScript keywords and messages lean on Twitch emote names. When a
new keyword, error message, or builtin needs an emote-flavored name,
propose candidates to the owner and wait for a pick before writing
them into the code. Do not invent emote references on your own; the
owner curates which memes count. (Established ones already in the
language, like PepoG or Wokege, need no approval when reused.)

## Verification

```sh
go test ./...
go vet ./...
gofmt -l .
```

All three must be clean before calling work done. The pty-based REPL
integration tests skip themselves off Linux.
