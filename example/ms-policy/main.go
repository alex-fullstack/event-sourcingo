package main

import (
	"os"
	"policy/cmd"
)

func main() {
	os.Exit(cmd.Execute())
}
