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
	return fmt.Sprintf("%s\n%s%s", lr.Name, indent+indent, strings.Join(lr.Labels, "\n"+indent+indent))
}

// Namespace represents a kubectl namespace. The name and the apps within it.
type Namespace struct {
	Name      string
	Resources []LabeledResource
}

// String satisfies the stringer interface.
func (n Namespace) String() string {
	if len(n.Resources) == 0 {
		return fmt.Sprintf("%s\n%s%s", color.Cyan(n.Name), indent, color.Yellow("(empty)"))
	}
	resourceStrings := make([]string, len(n.Resources))
	for i := range n.Resources {
		resourceStrings[i] = n.Resources[i].String()
	}
	return fmt.Sprintf("%s\n%s%s", color.Cyan(n.Name), indent, strings.Join(resourceStrings, "\n"+indent))
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

	if isServiceKind(kind) {
		return getServiceLabels(payload), nil
	}

	if isDeploymentKind(kind) {
		return getDeploymentLabels(payload), nil
	}

	md := payload["metadata"].(map[string]interface{})
	labels := md["labels"].(map[string]interface{})

	result := []string{}
	for key, value := range labels {
		result = append(result, key+"="+value.(string))
	}

	sort.SliceStable(result, func(i, j int) bool {
		return result[i] < result[j]
	})

	return result, nil
}

func getServiceLabels(definition map[string]interface{}) []string {
	spec := definition["spec"].(map[string]interface{})
	selector := spec["selector"].(map[string]interface{})
	result := []string{}
	for key, value := range selector {
		result = append(result, key+"="+value.(string))
	}
	return result
}

func getDeploymentLabels(definition map[string]interface{}) []string {
	spec := definition["spec"].(map[string]interface{})
	selector := spec["selector"].(map[string]interface{})
	matchLabels := selector["matchLabels"].(map[string]interface{})
	result := []string{}
	for key, value := range matchLabels {
		result = append(result, key+"="+value.(string))
	}
	return result
}

func isServiceKind(kind string) bool {
	return kind == "svc" || kind == "service" || kind == "services"
}

func isDeploymentKind(kind string) bool {
	return kind == "deploy" || kind == "deployment" || kind == "deployments"
}

// GetServicePods gets all podname for a label
func getPodsByLabel(n, label string) ([]string, error) {
	output, err := exec.Command("kubectl", "-n", n, "get", "pods", "--selector", label, "-o", `jsonpath={.items[*].metadata.name}`).Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get pods using label %s: %v", label, err)
	}
	if len(output) == 0 {
		return nil, nil
	}
	return strings.Split(string(output), " "), nil
}
