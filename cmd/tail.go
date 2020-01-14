package cmd

import (
	"github.com/davidmdm/kubelog/kubectl"
)

// Tail streams all pods for an application in a namespace to stdout
func Tail(namespace string, services []string, opts kubectl.LogOptions) error {
	c := make(chan error)

	if len(services) == 1 && services[0] == "*" {
		var err error
		services, err = kubectl.GetServicesByNamespace(namespace)
		if err != nil {
			return err
		}
	}

	for _, service := range services {
		go func(service string) {
			err := kubectl.TailLogs(namespace, service, opts)
			if err != nil {
				c <- err
			}
		}(service)
	}
	return <-c
}
