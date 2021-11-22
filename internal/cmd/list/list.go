package list

import (
	"context"
	"fmt"
	"strings"

	"github.com/davidmdm/kubelog/internal/cmd"
	"github.com/davidmdm/kubelog/internal/color"
	"github.com/davidmdm/kubelog/internal/kubectl"
	"github.com/davidmdm/kubelog/internal/terminal"
	"github.com/spf13/cobra"
)

// LogNamespace will log apps for a namespace. If an empty string is provided as namespace
// it will log all apps for all namespaces

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "ls",
		RunE: func(cmd *cobra.Command, args []string) error {
			namespace, err := cmd.Flags().GetString("namespace")
			if err != nil {
				return err
			}
			return listPods(cmd.Context(), namespace, args)
		},
	}

	cmd.Flags().StringP("namespace", "n", "", "kubectl namespace to use, if not provided will run for all namespaces")

	return cmd
}

func listPods(ctx context.Context, namespace string, filters []string) error {
	ctl, err := kubectl.NewCtl(namespace)
	if err != nil {
		return fmt.Errorf("failed to connect to kubernetes: %w", err)
	}

	if namespace == "" {
		namespace, err = cmd.SelectNamespace(ctx, ctl)
		if err != nil {
			return err
		}
		*ctl = ctl.WithNamespace(namespace)
	}

	pods, err := ctl.GetPods(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to list pods: %w", err)
	}

	for _, pod := range pods {
		if len(filters) > 0 && !matchesSubstrings(pod.Name, filters) {
			continue
		}

		terminal.Println(color.Cyan(pod.Name))
		for k, v := range pod.Labels {
			terminal.Printf("  %s=%s\n", strings.TrimSpace(k), strings.TrimSpace(v))
		}
		terminal.Println()
	}

	return nil
}

func matchesSubstrings(value string, filters []string) bool {
	for _, filter := range filters {
		if strings.Contains(value, filter) {
			return true
		}
	}
	return false
}
