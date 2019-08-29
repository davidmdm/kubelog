package kubectl

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
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
	out, err := exec.Command("kubectl", "get", "namespaces", "-o", "jsonpath='{.items[*].metadata.name}'").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubectl namespaces: %v", err)
	}
	return strings.Split(string(out[1:len(out)-1]), " "), nil
}

// GetPodsByNamespace returns all pods in a namespace
func GetPodsByNamespace(namespace, selector string) ([]string, error) {
	out, err := exec.Command("kubectl", "get", "pods", "-n", namespace, "--selector", "app="+selector, "-o", "jsonpath='{.items[*].metadata.name}'").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute kubectl get pods: %v", err)
	}
	return strings.Split(string(out[1:len(out)-1]), " "), nil
}

// GetServicesByNamespace will return the service names by namespace
func GetServicesByNamespace(name string) ([]string, error) {
	out, err := exec.Command("kubectl", "-n", name, "get", "services", "-o", "jsonpath='{.items[*].metadata.name}'").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get service names: %v", err)
	}
	return strings.Split(string(out[1:len(out)-1]), " "), nil
}
