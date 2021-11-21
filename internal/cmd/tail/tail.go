package tail

import (
	"bufio"
	"context"
	"fmt"
	"io"

	"github.com/davidmdm/kubelog/internal/kubectl"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/watch"

	corev1 "k8s.io/api/core/v1"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "tail [labels...]",
		RunE: func(cmd *cobra.Command, args []string) error {
			namespace, err := cmd.Flags().GetString("namespace")
			if err != nil {
				return err
			}

			return tail(cmd.Context(), namespace, args)
		},
	}

	cmd.MarkFlagRequired("namespace")
	cmd.Flags().StringP("namespace", "n", "", "kubectl namespace to use")
	cmd.Flags().StringP("prefix", "p", "", "prepends a prefix and equal sign to all passed input labels")
	cmd.Flags().StringP("since", "s", "", "kubectl since option for logs")
	cmd.Flags().BoolP("timestamp", "t", false, "kubectl timestamp option for logs")

	return cmd
}

func tail(ctx context.Context, namespace string, labels []string) error {
	ctl, err := kubectl.NewCtl()
	if err != nil {
		return fmt.Errorf("failed to connect to kubernetes: %w", err)
	}

	if len(labels) == 0 {
		labels = append(labels, "")
	}

	watchers := make([]<-chan kubectl.PodEvent, len(labels))
	for _, label := range labels {
		watcher, err := ctl.WatchPods(ctx, namespace, label)
		if err != nil {
			return fmt.Errorf("failed to watch pods: %w", err)
		}
		watchers = append(watchers, watcher)
	}

	podWatcher := JoinChannels(watchers...)

	pods := make(chan *corev1.Pod)

	go func() {
		for podEvent := range podWatcher {
			if podEvent.Type != watch.Added || podEvent.Pod == nil {
				continue
			}
			pods <- podEvent.Pod
		}
	}()

	output := make(chan string)
	go func() {
		for pod := range pods {
			for _, container := range pod.Spec.Containers {
				rc, err := ctl.StreamPodLogs(ctx, namespace, pod.Name, kubectl.PodLogOptions{
					Container:  container.Name,
					Follow:     true,
					Previous:   false,
					Timestamps: false,
					Since:      nil,
				})
				if err != nil {
					fmt.Printf("failed to stream logs for pod \"%s:%s\": %v\n", pod.Name, container.Name, err)
					continue
				}

				go func(rc io.ReadCloser) {
					defer rc.Close()
					scanner := bufio.NewScanner(rc)
					for scanner.Scan() {
						output <- scanner.Text()
					}
					if err := scanner.Err(); err != nil {

					}
				}(rc)
			}
		}
	}()

	return nil
}

func JoinChannels[T any](channels ...<-chan T) <-chan T {
	result := make(chan T)

	for _, c := range channels {
		go func(c <-chan T) {
			for t := range c {
				result <- t
			}
		}(c)
	}

	return result
}
