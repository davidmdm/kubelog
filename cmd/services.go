package cmd

import (
	"fmt"
	"sort"
	"sync"

	"github.com/davidmdm/kubelog/kubectl"
)

// LogNamespace will log apps for a namespace. If an empty string is provided as namespace
// it will log all apps for all namespaces
func LogNamespace(name string) error {
	var namespaceNames []string
	if name == "" {
		names, err := kubectl.GetNamespaceNames()
		if err != nil {
			return fmt.Errorf("failed to fetch namespaces: %v", err)
		}
		namespaceNames = names
	} else {
		namespaceNames = append(namespaceNames, name)
	}

	results := []*kubectl.Namespace{}
	errors := []error{}

	var wg sync.WaitGroup
	wg.Add(len(namespaceNames))

	var mu sync.Mutex
	for _, name := range namespaceNames {
		go func(name string) {
			services, err := kubectl.GetServicesByNamespace(name)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errors = append(errors, fmt.Errorf("error fetching namespace %s: %v", name, err))
			} else {
				results = append(results, &kubectl.Namespace{Name: name, Apps: services})
			}
			wg.Done()
		}(name)
	}

	wg.Wait()

	sort.SliceStable(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	for _, result := range results {
		fmt.Println(result)
		fmt.Println()
	}
	for _, err := range errors {
		fmt.Printf("error: %v\n", err)
	}

	return nil
}
