package kubectl

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"

	"github.com/davidmdm/kubelog/util/color"
)

var spaceRegex = regexp.MustCompile(`\s`)
var runningStatus = regexp.MustCompile(`\sRunning\s`)

// GetNamespaceNames returns all namespace for your kube config
func GetNamespaceNames() ([]string, error) {
	out, err := exec.Command("kubectl", "get", "namespaces").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubectl namespace: %v", err)
	}

	lines := bytes.Split(out, []byte("\n"))
	lines = lines[1 : len(lines)-1]

	namespaces := []string{}

	for _, line := range lines {
		if string(line) != "" {
			namespaces = append(namespaces, string(line[0:spaceRegex.FindIndex(line)[0]]))
		}
	}
	return namespaces, nil
}

// GetRunningPodsByNamespace returns all pods in a namespace
func GetRunningPodsByNamespace(namespace string) ([]string, error) {
	out, err := exec.Command("kubectl", "get", "pods", "-n", namespace).Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute kubectl get pods: %v", err)
	}

	lines := bytes.Split(out, []byte{'\n'})

	podLines := [][]byte{}
	for _, line := range lines {
		if runningStatus.Match(line) {
			podLines = append(podLines, line)
		}
	}

	pods := []string{}
	for _, line := range podLines {
		podName := line[0:spaceRegex.FindIndex(line)[0]]
		pods = append(pods, string(podName))
	}

	return pods, nil
}

type LogOptions struct {
	Timestamps bool
	Since      string
}

// FollowLog return a channel that gives you the strings line by line of a pods log
func FollowLog(namepace, pod string, activePods *PodList, opts LogOptions) error {
	args := []string{"logs", pod, "-f", "-n", namepace}
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
	prefix := color.Color(pod)

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
