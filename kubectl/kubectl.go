package kubectl

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
)

var spaceRegex = regexp.MustCompile(`\s`)

// GetNamespaceNames returns all namespace for your kube config
func GetNamespaceNames() ([][]byte, error) {
	out, err := exec.Command("kubectl", "get", "namespaces").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubectl namespace: %v", err)
	}

	lines := bytes.Split(out, []byte("\n"))
	lines = lines[1 : len(lines)-1]

	namespaces := [][]byte{}

	for _, line := range lines {
		if string(line) != "" {
			namespaces = append(namespaces, line[0:spaceRegex.FindIndex(line)[0]])
		}
	}
	return namespaces, nil
}

// GetPodsByNamespace returns all pods in a namespace
func GetPodsByNamespace(namespace string) ([][]byte, error) {
	out, err := exec.Command("kubectl", "get", "pods", "-n", namespace).Output()
	if (err) != nil {
		return nil, fmt.Errorf("failed to execute kubectl get pods: %v", err)
	}

	lines := bytes.Split(out, []byte{'\n'})
	lines = lines[:len(lines)-1]

	if len(lines) < 2 {
		lines = [][]byte{}
	} else {
		lines = lines[1:]
	}

	pods := [][]byte{}
	for _, line := range lines {
		podName := line[0:spaceRegex.FindIndex(line)[0]]
		pods = append(pods, podName)
	}

	return pods, nil
}
