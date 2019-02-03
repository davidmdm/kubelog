package kubectl

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
)

var spaceRegex = regexp.MustCompile(`\s`)
var podStatus = regexp.MustCompile(`Running|Terminating|CrashLoopBackoff`)

const (
	red     = 31
	green   = 32
	yellow  = 33
	blue    = 34
	magenta = 35
	cyan    = 36
)

var i int
var colors = []int{cyan, magenta, yellow, blue, red, green}

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

// GetPodsByNamespace returns all pods in a namespace
func GetPodsByNamespace(namespace string) ([]string, error) {
	out, err := exec.Command("kubectl", "get", "pods", "-n", namespace).Output()
	if (err) != nil {
		return nil, fmt.Errorf("failed to execute kubectl get pods: %v", err)
	}

	lines := bytes.Split(out, []byte{'\n'})

	podLines := [][]byte{}
	for _, line := range lines {
		if podStatus.Match(line) {
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

// FollowLog return a channel that gives you the strings line by line of a pods log
func FollowLog(namepace, pod string) (<-chan (string), error) {

	cmd := exec.Command("kubectl", "logs", pod, "-f", "-n", namepace)
	cmd.Stderr = os.Stderr

	rc, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to connect sub pipe: %v", err)
	}
	if err = cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start running kubectl: %v", err)
	}

	r := bufio.NewReader(rc)
	lines := make(chan string)
	prefix := color(pod)

	go func() {
		defer rc.Close()
		defer close(lines)
		for {
			l, err := r.ReadString('\n')
			if err != nil {
				// handle error somehow???
				return
			}
			lines <- prefix + "  " + l
		}
	}()

	return lines, nil
}

func color(str string) (ret string) {
	ret = fmt.Sprintf("\033[%d;3m%s\033[0m", colors[i], str)
	i = (i + 1) % len(colors)
	return
}
