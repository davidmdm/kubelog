package kubectl

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/davidmdm/kubelog/internal/util/color"
)

// LogOptions sets whether the logs should include a timestamp and how far back since now we need to fetch the logs.
// by default there are no timestamps and logs will be fetched since their beginning.
type LogOptions struct {
	Timestamps  bool
	Since       string
	LabelPrefix string
}

// TailLogs will start outputting all logs to stdout for a every pod in the given service for a specific namespace
func TailLogs(namespace, label string, opts LogOptions) {
	activePods := new(podList)
	monitorPods(activePods, namespace, label, opts)
	for range time.NewTicker(10 * time.Second).C {
		monitorPods(activePods, namespace, label, opts)
	}
}

func monitorPods(activePods *podList, namespace, label string, opts LogOptions) {
	pods, err := getPodsByLabel(namespace, label)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to fetch pods: %v\ntrying again in 10 seconds...\n", err)
		return
	}

	if len(pods) == 0 {
		fmt.Fprintf(os.Stderr, "There are no pods for label %s\n", label)
		return
	}
	for _, pod := range pods {
		for _, podcontainer := range pod.expand() {
			if activePods.has(podcontainer) {
				continue
			}
			if err := tailPod(activePods, namespace, podcontainer, opts); err != nil {
				fmt.Fprintf(os.Stderr, "failed to tail pod \"%s\": %v\n", pod, err)
			}
		}
	}
}

func tailPod(activePods *podList, namespace, pod string, opts LogOptions) error {
	options := []string{"-f", "-n", namespace}
	podSegments := strings.Split(pod, "/")

	args := append([]string{"logs"}, podSegments...)
	args = append(args, options...)

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
				if errors.Is(err, io.EOF) {
					fmt.Printf("%s  %s", prefix, line+" - EOF\n")
				} else {
					fmt.Fprintf(os.Stderr, "\nunexpected error reading log for pod %s: %v\n\n", pod, err)
				}
				activePods.remove(pod)
				return
			}
			fmt.Printf("%s  %s", prefix, line)
		}
	}()

	return nil
}
