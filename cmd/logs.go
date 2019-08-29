package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/davidmdm/kubelog/kubectl"
)

// StreamLogs streams all pods for an application in a namespace to stdout
func StreamLogs(namespace, service string, opts kubectl.LogOptions) {
	monitorPods(namespace, service, opts)
	for range time.NewTicker(10 * time.Second).C {
		monitorPods(namespace, service, opts)
	}
}

func monitorPods(namespace, service string, opts kubectl.LogOptions) {
	pods, err := kubectl.GetServicePods(namespace, service)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to fetch pods: %v\ntrying again in 10 seconds...\n", err)
		return
	}
	if len(pods) == 0 {
		fmt.Fprintf(os.Stderr, "There are no pods for service %s", service)
		return
	}
	for _, pod := range pods {
		if err := kubectl.FollowLog(namespace, pod, opts); err != nil {
			fmt.Fprintf(os.Stderr, "failed to follow log for pod %s: %v\n", pod, err)
		}
	}
}
