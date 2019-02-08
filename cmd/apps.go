package cmd

import (
	"fmt"
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

	results := make(chan *kubectl.Namespace)
	errors := make(chan error)
	done := make(chan (struct{}))

	go func() {
		for result := range results {
			fmt.Println(result)
			fmt.Println()
		}
		for err := range errors {
			fmt.Printf("\nerror: %v\n\n", err)
		}
		done <- struct{}{}
	}()

	var wg sync.WaitGroup
	wg.Add(len(namespaceNames))

	for _, name := range namespaceNames {
		go func(name string) {
			ns, err := kubectl.GetNamespace(name)
			if err != nil {
				errors <- fmt.Errorf("error fetching namespace %s: %v", name, err)
			} else {
				results <- ns
			}
			wg.Done()
		}(name)
	}

	wg.Wait()

	close(results)
	close(errors)

	<-done

	return nil
}
