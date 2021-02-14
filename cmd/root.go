package cmd

import (
	"github.com/davidmdm/kubelog/cmd/get"
	"github.com/davidmdm/kubelog/cmd/tail"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "kubelog",
		Short:        "view combined output logs for kubernetes services",
		Long:         `Kubelog is a CLI for quickly listing your services per namespace and viewing the combined logs of all pods running in those services`,
		SilenceUsage: true,
	}

	cmd.AddCommand(tail.Cmd())
	cmd.AddCommand(get.Cmd())

	return cmd
}
