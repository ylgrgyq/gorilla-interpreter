package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"repl"
)

func main() {
	modePtr := flag.String("mode", "compiler", "compiler or interpreter")

	flag.Parse()

	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Hello %s! This is the Gorilla programming language running in %q mode!\n", user.Username, *modePtr)

	fmt.Printf("Feel free to type in commands\n")

	if *modePtr == "compiler" {
		repl.StartWithCompiler(os.Stdin, os.Stdout)
	} else {
		repl.StartWithInterpreter(os.Stdin, os.Stdout)
	}

}
