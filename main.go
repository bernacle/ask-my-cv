package main

import (
	"os"

	"github.com/brunosimplicio/ask-my-cv/cmd"
)

func main() {
	os.Exit(cmd.Execute(os.Args[1:]))
}
