package main

import (
	"fmt"
	"os"

	"github.com/ohnishat/peepoScript/cmd/repl"
)

const PEEPO_SCRIPT_VERSION = "0.1.0"

func main() {
	fmt.Printf("Peepo Script - %s | REPL\n\n\n", PEEPO_SCRIPT_VERSION)
	repl.StartRepl(os.Stdin, os.Stdout)
}
