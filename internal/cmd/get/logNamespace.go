package get

import (
	"fmt"
	"sort"
	"sync"

	"github.com/davidmdm/kubelog/internal/kubectl"
	"github.com/spf13/cobra"
)

// LogNamespace will log apps for a namespace. If an empty string is provided as namespace
// it will log all apps for all namespaces

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "get [resource]",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			namespace, err := cmd.Flags().GetString("namespace")
			if err != nil {
				return err
			}
			return logNamespace(namespace, args[0])
		},
	}

	cmd.Flags().StringP("namespace", "n", "", "kubectl namespace to use, if not provided will run for all namespaces")

	return cmd
}

func logNamespace(ns, kind string) error {
	var namespaceNames []string
	if ns == "" {
		names, err := kubectl.GetNamespaceNames()
		if err != nil {
			return fmt.Errorf("failed to fetch namespaces: %v", err)
		}
		namespaceNames = names
	} else {
		namespaceNames = append(namespaceNames, ns)
	}

	results := []*kubectl.Namespace{}
	errors := []error{}

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, name := range namespaceNames {
		wg.Add(1)
		go func(name string) {
			services, err := kubectl.GetResourcesByNamespace(name, kind)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errors = append(errors, fmt.Errorf("error fetching namespace %s: %v", name, err))
			} else {
				results = append(results, &kubectl.Namespace{Name: name, Resources: services})
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
