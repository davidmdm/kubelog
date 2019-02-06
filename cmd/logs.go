package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/davidmdm/kubelog/kubectl"
)

// StreamLogs streams all pods for an application in a namespace to stdout
func StreamLogs(n, a string, timestamp bool, since string) {
	activePods := new(kubectl.PodList)
	monitorPods(n, a, activePods, timestamp, since)

	// here we want to purposefully block the thread forever as we continue monitoring in other goroutines
	<-make(chan struct{})
}

func monitorPods(n, a string, activePods *kubectl.PodList, timestamp bool, since string) {

	defer time.AfterFunc(10*time.Second, func() { monitorPods(n, a, activePods, timestamp, "") })

	appPods, err := getAppPods(n, a)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nfailed to fetch pods: %v\ntrying again in 10 seconds...\n\n", err)
		return
	}

	for _, pod := range appPods {
		if !activePods.Has(pod) {
			if err := kubectl.FollowLog(n, pod, activePods, timestamp, since); err != nil {
				fmt.Fprintf(os.Stderr, "\nfailed to follow log for pod %s: %v\n\n", pod, err)
			}
		}
	}

	if activePods.Length() == 0 {
		fmt.Printf("there are no active pods for `%s` in `%s`\n", a, n)
	}
}

func getAppPods(n, a string) ([]string, error) {
	pods, err := kubectl.GetRunningPodsByNamespace(n)
	if err != nil {
		return nil, fmt.Errorf("failed to get pods by namespace: %v", err)
	}
	appPods := []string{}
	for _, pod := range pods {
		if strings.HasPrefix(pod, a+"-") {
			appPods = append(appPods, pod)
		}
	}
	return appPods, nil
}
