package cmd

import (
	"fmt"
	"strings"
	"sync"

	"github.com/davidmdm/kubelog/kubectl"
)

type namespace struct {
	name string
	apps [][]byte
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
				fmt.Println(string(app))
			}
			fmt.Println()
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
					errors <- err
				} else {
					results <- ns
				}
				wg.Done()
			}(string(name))
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

func getAppsByNamespace(namespace string) ([][]byte, error) {
	pods, err := kubectl.GetPodsByNamespace(namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get pods: %v", err)
	}

	apps := [][]byte{}
	for _, pod := range pods {
		app := getAppFromPodName(pod)
		if !contains(apps, app) {
			apps = append(apps, app)
		}
	}

	return apps, nil
}

func getAppFromPodName(pod []byte) []byte {
	idx := []int{}
	for i, b := range pod {
		if b == '-' {
			idx = append(idx, i)
		}
	}

	if len(idx) < 2 {
		return pod
	}

	return pod[:idx[len(idx)-2]]
}

func contains(set [][]byte, elem []byte) bool {
	for _, value := range set {
		if string(elem) == string(value) {
			return true
		}
	}
	return false
}
