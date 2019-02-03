package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/davidmdm/kubelog/cmd"
)

func main() {
	namespace := flag.String("n", "", "namespace")

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
		if err := cmd.StreamLogs(*namespace, args[0]); err != nil {
			fmt.Fprintf(os.Stderr, "failed to stream logs: %v", err)
		}
		return
	}

	fmt.Fprintf(os.Stderr, "command not recognized. Available commands are `get apps` or `[app]`\n")
}
