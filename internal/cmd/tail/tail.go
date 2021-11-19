package tail

import (
	"fmt"

	"github.com/davidmdm/kubelog/internal/kubectl"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
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
			labelPrefix, err := cmd.Flags().GetString("prefix")
			if err != nil {
				return err
			}
			return tail(namespace, args, kubectl.LogOptions{Timestamps: timestamp, Since: since, LabelPrefix: labelPrefix})
		},
	}

	cmd.MarkFlagRequired("namespace")
	cmd.Flags().StringP("namespace", "n", "", "kubectl namespace to use")
	cmd.Flags().StringP("prefix", "p", "", "prepends a prefix and equal sign to all passed input labels")
	cmd.Flags().StringP("since", "s", "", "kubectl since option for logs")
	cmd.Flags().BoolP("timestamp", "t", false, "kubectl timestamp option for logs")

	return cmd
}

func tail(namespace string, labels []string, opts kubectl.LogOptions) error {
	for _, label := range labels {
		if opts.LabelPrefix != "" {
			label = fmt.Sprintf("%s=%s", opts.LabelPrefix, label)
		}
		go kubectl.TailLogs(namespace, label, opts)
	}

	// at this point we never want to return since we want to monitor the logs forever
	select {}
}
