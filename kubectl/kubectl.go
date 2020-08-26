package kubectl

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"sync"

	"github.com/davidmdm/kubelog/util/color"
)

const indent = "  "

type LabeledResource struct {
	Name   string
	Labels []string
}

func (lr LabeledResource) String() string {
	return fmt.Sprintf("%s   %s", lr.Name, strings.Join(lr.Labels, " "))
}

// Namespace represents a kubectl namespace. The name and the apps within it.
type Namespace struct {
	Name        string
	Services    []LabeledResource
	Deployments []LabeledResource
}

// String satisfies the stringer interface.
func (n Namespace) String() string {
	if len(n.Services) == 0 {
		return fmt.Sprintf("%s\n%s%s", color.Cyan(n.Name), indent, color.Yellow("(empty)"))
	}
	serviceStrings := make([]string, len(n.Services))
	for i := range n.Services {
		serviceStrings[i] = n.Services[i].String()
	}
	return fmt.Sprintf("%s\n%s%s", color.Cyan(n.Name), indent, strings.Join(serviceStrings, "\n"+indent))
}

// GetNamespaceNames returns all namespace for your kube config
func GetNamespaceNames() ([]string, error) {
	out, err := exec.Command("kubectl", "get", "namespaces", "-o", "jsonpath={.items[*].metadata.name}").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubectl namespaces: %v", err)
	}
	return strings.Split(string(out), " "), nil
}

// GetResourcesByNamespace will return the service names by namespace
func GetResourcesByNamespace(ns, kind string) ([]LabeledResource, error) {
	out, err := exec.Command("kubectl", "-n", ns, "get", kind, "-o", "jsonpath={.items[*].metadata.name}").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get service names: %v", err)
	}

	resourceNames := []string{}
	for _, str := range strings.Split(string(out), " ") {
		if str != "" {
			resourceNames = append(resourceNames, str)
		}
	}

	wg := new(sync.WaitGroup)
	wg.Add(len(resourceNames))
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	resources := make([]LabeledResource, len(resourceNames))
	errChan := make(chan error, len(resourceNames))
	for i := range resourceNames {
		go func(idx int, name string) {
			defer wg.Done()
			if svc, err := getResource(ns, kind, name); err != nil {
				errChan <- err
			} else {
				resources[idx] = *svc
			}
		}(i, resourceNames[i])
	}

	select {
	case <-done:
		sort.SliceStable(resources, func(i, j int) bool {
			return resources[i].Name < resources[j].Name
		})
		return resources, nil
	case err := <-errChan:
		return nil, err
	}

}

func getResource(ns, kind, serviceName string) (*LabeledResource, error) {
	labels, err := getResourceLabels(ns, kind, serviceName)
	if err != nil {
		return nil, err
	}
	return &LabeledResource{Name: serviceName, Labels: labels}, nil
}

func getResourceLabels(ns, kind, id string) ([]string, error) {
	out, err := exec.Command("kubectl", "-n", ns, "get", kind, id, "-o", "json").Output()
	if err != nil {
		return nil, err
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(out, &payload); err != nil {
		return nil, err
	}

	md := payload["metadata"].(map[string]interface{})
	labels := md["labels"].(map[string]string)

	result := []string{}
	for key, value := range labels {
		result = append(result, key+"="+value)
	}

	sort.SliceStable(result, func(i, j int) bool {
		return result[i] < result[j]
	})

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
