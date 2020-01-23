package cmd

import (
	"github.com/davidmdm/kubelog/kubectl"
)

// Tail streams all pods for an application in a namespace to stdout
func Tail(namespace string, services []string, opts kubectl.LogOptions) error {
	if len(services) == 1 && services[0] == "*" {
		svcs, err := kubectl.GetServicesByNamespace(namespace)
		if err != nil {
			return err
		}
		services = svcs
	}

	for _, service := range services {
		go kubectl.TailLogs(namespace, service, opts)
	}

	// at this point we never want to return since we want to monitor the logs forever
	select {}
}
