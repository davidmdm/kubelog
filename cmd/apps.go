package cmd

import (
	"fmt"
	"strings"
	"sync"

	"github.com/davidmdm/kubelog/kubectl"
)

type namespace struct {
	name string
	apps []string
}

// LogNamespace will log apps for a namespace. If an empty string is provided as namespace
// it will log all apps for all namespaces
func LogNamespace(name string) error {
	results := make(chan *namespace)
	errors := make(chan error)
	done := make(chan (struct{}))

	go func() {
		for result := range results {
			fmt.Println(strings.ToUpper(result.name))
			for _, app := range result.apps {
				fmt.Println(app)
			}
			fmt.Println()
		}
		for err := range errors {
			fmt.Printf("\nerror: %v\n\n", err)
		}
		done <- struct{}{}
	}()

	if name == "" {
		names, err := kubectl.GetNamespaceNames()
		if err != nil {
			return fmt.Errorf("failed to fetch namespaces: %v", err)
		}

		var wg sync.WaitGroup
		wg.Add(len(names))

		for _, name := range names {
			go func(name string) {
				ns, err := getNamespace(name)
				if err != nil {
					errors <- fmt.Errorf("error fetching namespace %s: %v", name, err)
				} else {
					results <- ns
				}
				wg.Done()
			}(name)
		}

		wg.Wait()
	} else {
		ns, err := getNamespace(name)
		if err != nil {
			errors <- err
		} else {
			results <- ns
		}
	}

	close(results)
	close(errors)

	<-done

	return nil
}

func getNamespace(name string) (*namespace, error) {
	apps, err := getAppsByNamespace(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get apps in namespace: %v", err)
	}
	return &namespace{name: name, apps: apps}, nil
}

func getAppsByNamespace(namespace string) ([]string, error) {
	pods, err := kubectl.GetPodsByNamespace(namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get pods: %v", err)
	}

	apps := []string{}
	for _, pod := range pods {
		app := getAppFromPodName(pod)
		if !contains(apps, app) {
			apps = append(apps, app)
		}
	}

	return apps, nil
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

func contains(set []string, elem string) bool {
	for _, value := range set {
		if elem == value {
			return true
		}
	}
	return false
}
