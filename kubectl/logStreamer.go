package kubectl

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"

	"github.com/davidmdm/kubelog/util/color"
)

// LogOptions sets whether the logs should include a timestamp and how far back since now we need to fetch the logs.
// by default there are no timestamps and logs will be fetched since their beginning.
type LogOptions struct {
	Timestamps bool
	Since      string
}

var activePods = PodList{}

// TailLogs return a channel that gives you the strings line by line of a pods log
func TailLogs(namespace, pod string, opts LogOptions) error {
	if activePods.Has(pod) {
		return nil
	}

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

	activePods.Add(pod)

	r := bufio.NewReader(rc)
	prefix := color.Color(namespace + "/" + pod)

	go func() {
		defer rc.Close()
		for {
			line, err := r.ReadString('\n')
			if err != nil {
				fmt.Fprintf(os.Stderr, "\nunexpected error reading log for pod %s: %v\n\n", pod, err)
				activePods.Remove(pod)
				return
			}
			fmt.Printf("%s  %s", prefix, line)
		}
	}()

	return nil
}
