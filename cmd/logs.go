package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/davidmdm/kubelog/kubectl"
)

// StreamLogs streams all pods for an application in a namespace to stdout
func StreamLogs(n, a string, opts kubectl.LogOptions) {
	monitorPods(n, a, new(kubectl.PodList), opts)

	// here we want to purposefully block the thread forever as we continue monitoring in other goroutines
	<-make(chan struct{})
}

func monitorPods(n, a string, activePods *kubectl.PodList, opts kubectl.LogOptions) {

	defer time.AfterFunc(10*time.Second, func() { monitorPods(n, a, activePods, kubectl.LogOptions{Timestamps: opts.Timestamps}) })

	appPods, err := getServicePods(n, a)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to fetch pods: %v\ntrying again in 10 seconds...\n", err)
		return
	}

	for _, pod := range appPods {
		if !activePods.Has(pod) {
			if err := kubectl.FollowLog(n, pod, activePods, opts); err != nil {
				fmt.Fprintf(os.Stderr, "failed to follow log for pod %s: %v\n", pod, err)
			}
		}
	}

	if activePods.Length() == 0 {
		fmt.Printf("there are no active pods for `%s` in `%s`\n", a, n)
	}
}

func getServicePods(n, serviceName string) ([]string, error) {

	selector, err := exec.Command("kubectl", "-n", n, "get", "svc", serviceName, "-o", "jsonpath='{.spec.selector.app}'").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get service selector: %v", err)
	}

	pods, err := kubectl.GetPodsByNamespace(n, string(selector[1:len(selector)-1]))
	if err != nil {
		return nil, fmt.Errorf("failed to get pods by namespace: %v", err)
	}

	return pods, nil
}
