package kubectl

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/davidmdm/kubelog/util/color"
)

// LogOptions sets whether the logs should include a timestamp and how far back since now we need to fetch the logs.
// by default there are no timestamps and logs will be fetched since their beginning.
type LogOptions struct {
	Timestamps bool
	Since      string
}

// TailLogs return a channel that gives you the strings line by line of a pods log
func TailLogs(namespace, service string, opts LogOptions) error {
	activePods := new(podList)

	monitorPods := func() {
		pods, err := GetServicePods(namespace, service)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to fetch pods: %v\ntrying again in 10 seconds...\n", err)
			return
		}
		if len(pods) == 0 {
			fmt.Fprintf(os.Stderr, "There are no pods for service %s", service)
			return
		}
		for _, pod := range pods {
			if activePods.has(pod) {
				continue
			}
			if err := tailPod(namespace, pod, activePods, opts); err != nil {
				fmt.Fprintf(os.Stderr, "failed to tail pod %s: %v\n", pod, err)
			}
		}
	}

	monitorPods()
	for range time.NewTicker(10 * time.Second).C {
		monitorPods()
	}

	return nil
}

func tailPod(namespace, pod string, activePods *podList, opts LogOptions) error {
	args := []string{"logs", pod, "-f", "-n", namespace}
	if opts.Timestamps {
		args = append(args, "--timestamps")
	}
	if opts.Since != "" {
		args = append(args, fmt.Sprintf("--since=%s", opts.Since))
	}

	cmd := exec.Command("kubectl", args...)

	rc, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to connect sub pipe: %v", err)
	}
	if err = cmd.Start(); err != nil {
		return fmt.Errorf("failed to start running kubectl: %v", err)
	}

	activePods.add(pod)

	r := bufio.NewReader(rc)
	prefix := color.Color(namespace + "/" + pod)

	go func() {
		defer rc.Close()
		for {
			line, err := r.ReadString('\n')
			if err != nil {
				fmt.Fprintf(os.Stderr, "\nunexpected error reading log for pod %s: %v\n\n", pod, err)
				activePods.remove(pod)
				return
			}
			fmt.Printf("%s  %s", prefix, line)
		}
	}()

	return nil
}
