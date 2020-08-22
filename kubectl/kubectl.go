package kubectl

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/davidmdm/kubelog/util/color"
)

const indent = "  "

// Namespace represents a kubectl namespace. The name and the apps within it.
type Namespace struct {
	Name     string
	Services []string
}

// String satisfies the stringer interface.
func (n Namespace) String() string {
	if len(n.Services) == 0 {
		return fmt.Sprintf("%s\n%s%s", color.Cyan(n.Name), indent, color.Yellow("(empty)"))
	}
	return fmt.Sprintf("%s\n%s%s", color.Cyan(n.Name), indent, strings.Join(n.Services, "\n"+indent))
}

// GetNamespaceNames returns all namespace for your kube config
func GetNamespaceNames() ([]string, error) {
	out, err := exec.Command("kubectl", "get", "namespaces", "-o", "jsonpath={.items[*].metadata.name}").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubectl namespaces: %v", err)
	}
	return strings.Split(string(out), " "), nil
}

// GetServicesByNamespace will return the service names by namespace
func GetServicesByNamespace(name string) ([]string, error) {
	out, err := exec.Command("kubectl", "-n", name, "get", "services", "-o", "jsonpath={.items[*].metadata.name}").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get service names: %v", err)
	}
	result := []string{}
	for _, str := range strings.Split(string(out), " ") {
		if str != "" {
			result = append(result, str)
		}
	}
	return result, nil
}

// GetServicePods gets all podname for a label
func getPodsByLabel(n, label string) ([]string, error) {
	output, err := exec.Command("kubectl", "-n", n, "get", "pods", "-l", label, "-o", `jsonpath={.items[*].metadata.name}`).Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get pods using label %s: %v", label, err)
	}
	return strings.Split(string(output), " "), nil
}
