package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/davidmdm/kubelog/cmd"
)

func main() {
	namespace := flag.String("n", "", "namespace")
	timestamp := flag.Bool("t", false, "enables timestamps for logs")

	flag.Parse()
	args := flag.Args()

	if len(args) == 2 && args[0] == "get" && (args[1] == "apps" || args[1] == "app") {
		if err := cmd.LogNamespace(*namespace); err != nil {
			fmt.Fprintf(os.Stderr, "failed to get apps: %v\n", err)
		}
		return
	}

	if len(args) == 1 {
		if *namespace == "" {
			fmt.Fprintf(os.Stderr, "namespace required\n")
			return
		}
		cmd.StreamLogs(strings.ToLower(*namespace), args[0], *timestamp)
		return
	}

	fmt.Fprintf(os.Stderr, "command not recognized. Available commands are `get apps` or `[app]`\n")
}
