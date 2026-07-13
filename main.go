package main

import (
	"fmt"
	"io"
	"os"

	"github.com/JuanCarlosAcostaPeraba/sparks-cli/cmd"
)

func main() {
	os.Exit(run(cmd.Execute, os.Stderr))
}

func run(execute func() error, errOut io.Writer) int {
	if err := execute(); err != nil {
		_, _ = fmt.Fprintln(errOut, err)
		return 1
	}
	return 0
}
