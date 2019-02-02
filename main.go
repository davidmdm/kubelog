package main

import (
	"flag"
	"log"

	"github.com/davidmdm/kubelog/cmd"
)

func main() {
	namespace := flag.String("n", "", "namespace")

	flag.Parse()

	err := cmd.LogNamespace(*namespace)
	if err != nil {
		log.Fatalf("program failed: %v", err)
	}

}
