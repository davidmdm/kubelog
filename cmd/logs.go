package cmd

import (
	"fmt"
	"strings"

	"github.com/davidmdm/kubelog/kubectl"
)

// StreamLogs streams all pods for an application in a namespace to stdout
func StreamLogs(n, a string) error {
	pods, err := kubectl.GetPodsByNamespace(n)
	if err != nil {
		return fmt.Errorf("failed to get pods by namespace: %v", err)
	}

	appPods := []string{}
	for _, pod := range pods {
		if strings.HasPrefix(pod, a) {
			appPods = append(appPods, pod)
		}
	}

	logs := []<-chan string{}

	for _, pod := range appPods {
		log, err := kubectl.FollowLog(n, pod)
		if err != nil {
			return fmt.Errorf("failed to follow log: %v", err)
		}
		logs = append(logs, log)
	}

	for line := range merge(logs...) {
		fmt.Print(line)
	}

	return nil
}

func merge(channels ...<-chan string) <-chan string {
	out := make(chan string)
	for _, c := range channels {
		go func(c <-chan string) {
			for v := range c {
				out <- v
			}
		}(c)
	}
	return out
}
