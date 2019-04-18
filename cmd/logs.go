package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"
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

	appPods, err := getAppPods(n, a)
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

func getAppPods(n, a string) ([]string, error) {
	pods, err := kubectl.GetPodsByNamespace(n)
	if err != nil {
		return nil, fmt.Errorf("failed to get pods by namespace: %v", err)
	}

	r, err := regexp.Compile("^" + strings.Replace(a, "*", `\w*`, -1) + "-")
	if err != nil {
		return nil, fmt.Errorf("failed to compile application regex: %v", err)
	}

	appPods := []string{}

	for _, pod := range pods {
		if r.Match([]byte(pod)) {
			appPods = append(appPods, pod)
		}
	}
	return appPods, nil
}
