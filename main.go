package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/davidmdm/kubelog/cmd"
	"github.com/davidmdm/kubelog/kubectl"
)

func main() {
	namespace := flag.String("n", "", "namespace")
	timestamp := flag.Bool("t", false, "enables timestamps for logs")
	since := flag.String("s", "", "get logs since how many seconds")

	flag.Parse()
	args := flag.Args()

	if len(args) == 2 && args[0] == "get" && (args[1] == "svc" || args[1] == "services") {
		if err := cmd.LogNamespace(*namespace); err != nil {
			fmt.Fprintf(os.Stderr, "failed to get services: %v\n", err)
			os.Exit(2)
		}
		return
	}

	if len(args) == 1 {
		if *namespace == "" {
			fmt.Fprintf(os.Stderr, "namespace required\n")
			os.Exit(1)
		}
		cmd.StreamLogs(*namespace, args[0], kubectl.LogOptions{Timestamps: *timestamp, Since: *since})
		return
	}

	fmt.Fprintf(os.Stderr, "command not recognized. Available commands are `get apps` or `[app]`\n")
	os.Exit(1)
}
