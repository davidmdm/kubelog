package cmd

import (
	"context"
	"fmt"

	"github.com/davidmdm/kubelog/internal/kubectl"
	"github.com/davidmdm/kubelog/internal/terminal"
)

func SelectNamespace(ctx context.Context, ctl *kubectl.K8Ctl) (string, error) {
	namespaces, err := ctl.GetNamespaces(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list namespaces: %w", err)
	}

	names := make([]string, len(namespaces))
	for i, namespace := range namespaces {
		names[i] = namespace.Name
	}

	name, err := terminal.Select("select namespace", names)
	if err != nil {
		err = fmt.Errorf("failed to select namespace: %w", err)
	}

	return name, err
}
