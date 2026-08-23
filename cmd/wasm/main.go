//go:build js && wasm

// Command wasm exposes the peepoScript interpreter to a browser page.
//
// It registers one global: peepoRun(code) -> {ok, output, value, errors}.
// peepoChat output is captured per call; the environment lives for the
// lifetime of the module, so bindings persist across runs.
package main

import (
	"sort"
	"strings"
	"syscall/js"

	"github.com/ohnishat/peepoScript/cmd/evaluator"
	"github.com/ohnishat/peepoScript/cmd/runner"
	"github.com/ohnishat/peepoScript/cmd/token"
)

func main() {
	out := &strings.Builder{}
	env := evaluator.NewEnvironmentWithOut(out)

	run := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) != 1 {
			return js.ValueOf(map[string]any{"ok": false, "errors": []any{"peepoRun expects exactly one argument"}})
		}
		out.Reset()
		res := runner.Run(args[0].String(), env)

		errs := make([]any, len(res.Errors))
		for i, e := range res.Errors {
			errs[i] = e
		}
		return js.ValueOf(map[string]any{
			"ok":     res.OK,
			"output": out.String(),
			"value":  res.Value,
			"errors": errs,
		})
	})
	defer run.Release()
	js.Global().Set("peepoRun", run)

	// Language symbols: keywords plus builtins. The page uses this one
	// list for both tab completion and emote lookup.
	keywords := js.FuncOf(func(this js.Value, args []js.Value) any {
		words := append(token.Keywords(), evaluator.Builtins()...)
		sort.Strings(words)
		arr := make([]any, len(words))
		for i, w := range words {
			arr[i] = w
		}
		return js.ValueOf(arr)
	})
	defer keywords.Release()
	js.Global().Set("peepoKeywords", keywords)

	// Signal readiness where a DOM is present; the worker host instead
	// relies on the globals above existing after instantiation.
	dispatch := js.Global().Get("dispatchEvent")
	if !dispatch.IsUndefined() {
		ready := js.Global().Get("CustomEvent").New("peepoReady")
		js.Global().Call("dispatchEvent", ready)
	}
	select {}
}
