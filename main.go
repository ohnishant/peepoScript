package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/ohnishat/peepoScript/cmd/evaluator"
	"github.com/ohnishat/peepoScript/cmd/repl"
	"github.com/ohnishat/peepoScript/cmd/runner"
)

const PEEPO_SCRIPT_VERSION = "0.1.0"

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

// run separates exit codes from the process so tests can exercise the
// whole CLI without spawning binaries.
func run(args []string, in io.Reader, out, errOut io.Writer) int {
	var inline string
	flags := flag.NewFlagSet("peepo", flag.ContinueOnError)
	flags.SetOutput(errOut)
	flags.StringVar(&inline, "c", "", "run peepoScript source directly and exit")
	flags.Usage = func() {
		fmt.Fprintf(errOut, "usage: peepo [-c code] [script.%s]\n\n", runner.DefaultExtension)
		fmt.Fprintln(errOut, "With no arguments the REPL starts. A script path without")
		fmt.Fprintf(errOut, "an extension gets %s appended when the bare path is missing.\n", runner.DefaultExtension)
		flags.PrintDefaults()
	}
	if err := flags.Parse(args); err != nil {
		return 2
	}

	env := evaluator.NewEnvironmentWithOut(out)

	if inline != "" {
		return execSource(inline, env, out, errOut)
	}

	switch flags.NArg() {
	case 0:
		fmt.Fprintf(out, "Peepo Script - %s | REPL\n\n\n", PEEPO_SCRIPT_VERSION)
		repl.StartRepl(in, out)
		return 0
	case 1:
		path, err := runner.ResolvePath(flags.Arg(0))
		if err != nil {
			fmt.Fprintln(errOut, err)
			return 1
		}
		source, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintln(errOut, err)
			return 1
		}
		return execSource(string(source), env, out, errOut)
	default:
		fmt.Fprintln(errOut, "peepo takes at most one script file")
		return 2
	}
}

func execSource(code string, env *evaluator.Environment, out, errOut io.Writer) int {
	_, err := runner.Run(code, env)
	if err != nil {
		fmt.Fprintln(errOut, err)
		return 1
	}
	return 0
}
