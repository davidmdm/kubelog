package kubectl

import (
	"fmt"
	"strings"

	"github.com/davidmdm/kubelog/util"
)

const indent = "  "

// Namespace represents a kubectl namespace. The name and the apps within it.
type Namespace struct {
	name string
	apps []string
}

// String satisfies the stringer interface.
func (n Namespace) String() string {
	return fmt.Sprintf("%s\n%s%s", n.name, indent, strings.Join(n.apps, "\n"+indent))
}

// GetNamespace returns a namespace for a specified namespace name.
func GetNamespace(name string) (*Namespace, error) {
	pods, err := GetRunningPodsByNamespace(name)
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

	return &Namespace{name: name, apps: apps}, nil
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
