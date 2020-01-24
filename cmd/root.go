package cmd

import "github.com/spf13/cobra"

import "github.com/davidmdm/kubelog/cmd/tail"

import "github.com/davidmdm/kubelog/cmd/get"

var rootCmd = &cobra.Command{
	Use:   "kubelog",
	Short: "view combined output logs for kubernetes services",
	Long:  `Kubelog is a CLI for quickly listing your services per namespace and viewing the combined logs of all pods running in those services`,
}

// Execute runs root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(tail.TailCmd, get.GetCommand)
}
