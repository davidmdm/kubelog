package main

import (
	"fmt"
	"os"

	"github.com/davidmdm/kubelog/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
