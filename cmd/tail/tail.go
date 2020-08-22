package tail

import (
	"github.com/davidmdm/kubelog/kubectl"
	"github.com/spf13/cobra"
)

// TailCmd is the command to tail
var TailCmd = &cobra.Command{
	Use:  "tail [labels...]",
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

func tail(namespace string, labels []string, opts kubectl.LogOptions) error {
	for _, label := range labels {
		go kubectl.TailLogs(namespace, label, opts)
	}

	// at this point we never want to return since we want to monitor the logs forever
	select {}
}
