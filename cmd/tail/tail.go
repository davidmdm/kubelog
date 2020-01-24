package tail

import (
	"github.com/davidmdm/kubelog/kubectl"
	"github.com/spf13/cobra"
)

// TailCmd is the command to tail
var TailCmd = &cobra.Command{
	Use:  "tail [services...]",
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		namespace, err := cmd.Flags().GetString("namespace")
		if err != nil {
			return err
		}
		since, err := cmd.Flags().GetString("since")
		if err != nil {
			return err
		}
		timestamp, err := cmd.Flags().GetBool("timestamp")
		if err != nil {
			return err
		}
		return tail(namespace, args, kubectl.LogOptions{Timestamps: timestamp, Since: since})
	},
}

func init() {
	TailCmd.Flags().StringP("namespace", "n", "", "kubectl namespace to use")
	TailCmd.MarkFlagRequired("namespace")

	TailCmd.Flags().StringP("since", "s", "", "kubectl since option for logs")
	TailCmd.Flags().BoolP("timestamp", "t", false, "kubectl timestamp option for logs")
}

func tail(namespace string, services []string, opts kubectl.LogOptions) error {
	if len(services) == 1 && services[0] == "*" {
		svcs, err := kubectl.GetServicesByNamespace(namespace)
		if err != nil {
			return err
		}
		services = svcs
	}

	for _, service := range services {
		go kubectl.TailLogs(namespace, service, opts)
	}

	// at this point we never want to return since we want to monitor the logs forever
	select {}
}
