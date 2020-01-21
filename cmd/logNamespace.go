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

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, name := range namespaceNames {
		wg.Add(1)
		go func(name string) {
			services, err := kubectl.GetServicesByNamespace(name)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errors = append(errors, fmt.Errorf("error fetching namespace %s: %v", name, err))
			} else {
				results = append(results, &kubectl.Namespace{Name: name, Services: services})
			}
			wg.Done()
		}(name)
	}

	wg.Wait()

	sort.SliceStable(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	for i := range results {
		fmt.Println(results[i])
		if i < len(results)-1 {
			fmt.Println()
		}
	}

	for _, err := range errors {
		fmt.Printf("%v\n", err)
	}

	return nil
}
