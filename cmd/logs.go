package cmd

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/davidmdm/kubelog/kubectl"
)

// StreamLogs streams all pods for an application in a namespace to stdout
func StreamLogs(n, a string, timestamp bool) {
	activePods := []string{}
	monitorPods(n, a, activePods, timestamp)

	// here we want to purposefully block the thread forever as we continue monitoring in other goroutines
	<-make(chan struct{})
}

func monitorPods(n, a string, activePods []string, timestamp bool) {

	defer time.AfterFunc(10*time.Second, func() { monitorPods(n, a, activePods, timestamp) })

	appPods, err := getAppPods(n, a)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nfailed to fetch pods: %v\ntrying again in 10 seconds...\n\n", err)
		return
	}

	logs := []<-chan string{}
	for _, pod := range appPods {
		if !contains(activePods, pod) {
			log, err := kubectl.FollowLog(n, pod, timestamp)
			if err != nil {
				fmt.Fprintf(os.Stderr, "\nfailed to follow log for pod %s: %v\n\n", pod, err)
			} else {
				logs = append(logs, log)
				activePods = append(activePods, pod)
			}
		}
	}

	inactivePodIdxs := []int{}
	for i, pod := range activePods {
		if !contains(appPods, pod) {
			inactivePodIdxs = append(inactivePodIdxs, i)
		}
	}
	for i := len(inactivePodIdxs) - 1; i > -1; i-- {
		activePods = append(activePods[:inactivePodIdxs[i]], activePods[inactivePodIdxs[i]+1:]...)
	}

	if len(activePods) == 0 {
		fmt.Printf("there are no active pods for `%s` in `%s`\n", a, n)
	}

	go func() {
		for line := range merge(logs...) {
			fmt.Print(line)
		}
	}()
}

func getAppPods(n, a string) ([]string, error) {
	pods, err := kubectl.GetPodsByNamespace(n)
	if err != nil {
		return nil, fmt.Errorf("failed to get pods by namespace: %v", err)
	}
	appPods := []string{}
	for _, pod := range pods {
		if strings.HasPrefix(pod, a) {
			appPods = append(appPods, pod)
		}
	}
	return appPods, nil
}

func merge(channels ...<-chan string) <-chan string {
	out := make(chan string)
	var wg sync.WaitGroup
	wg.Add(len(channels))
	for _, c := range channels {
		go func(c <-chan string) {
			for v := range c {
				out <- v
			}
			wg.Done()
		}(c)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
