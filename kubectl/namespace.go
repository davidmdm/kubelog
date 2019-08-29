package kubectl

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/davidmdm/kubelog/util"
)

const indent = "  "

var spaceRegex = regexp.MustCompile(`\s`)
var podStatus = regexp.MustCompile(`\s(Running|CrashLoopBackOff|Error)\s`)

// Namespace represents a kubectl namespace. The name and the apps within it.
type Namespace struct {
	Name string
	Apps []string
}

// String satisfies the stringer interface.
func (n Namespace) String() string {
	return fmt.Sprintf("%s\n%s%s", n.Name, indent, strings.Join(n.Apps, "\n"+indent))
}

// GetNamespaceNames returns all namespace for your kube config
func GetNamespaceNames() ([]string, error) {
	out, err := exec.Command("kubectl", "get", "namespaces").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubectl namespaces: %v", err)
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
	if err != nil {
		return nil, fmt.Errorf("failed to execute kubectl get pods: %v", err)
	}

	lines := bytes.Split(out, []byte{'\n'})
	pods := []string{}
	for _, line := range lines {
		if podStatus.Match(line) {
			podName := line[0:spaceRegex.FindIndex(line)[0]]
			pods = append(pods, string(podName))
		}
	}

	return pods, nil
}

// GetNamespace returns a namespace for a specified namespace name.
func GetNamespace(name string) (*Namespace, error) {
	pods, err := GetPodsByNamespace(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get pods: %v", err)
	}

	apps := []string{}
	for _, pod := range pods {
		app := getAppFromPodName(pod)
		if !util.HasString(apps, app) {
			apps = append(apps, app)
		}
	}

	return &Namespace{Name: name, Apps: apps}, nil
}

// GetServicesByNamespace will return the service names by namespace
func GetServicesByNamespace(name string) ([]string, error) {
	out, err := exec.Command("kubectl", "-n", name, "get", "services", "-o", "jsonpath='{.items[*].metadata.name}'").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get service names: %v", err)
	}
	return strings.Split(string(out[1:len(out)-1]), " "), nil
}

func getAppFromPodName(pod string) string {
	idx := []int{}
	for i, b := range []byte(pod) {
		if b == '-' {
			idx = append(idx, i)
		}
	}
	if len(idx) < 2 {
		return pod
	}
	return pod[:idx[len(idx)-2]]
}
