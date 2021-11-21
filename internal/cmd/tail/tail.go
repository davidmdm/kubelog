package tail

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/davidmdm/kubelog/internal/kubectl"

	"github.com/davidmdm/kubelog/internal/color"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/wait"
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

			rawSince, err := cmd.Flags().GetString("since")
			if err != nil {
				return err
			}

			var since *time.Duration
			if rawSince != "" {
				since = new(time.Duration)
				*since, err = time.ParseDuration(rawSince)
				if err != nil {
					return err
				}
			}

			timestamp, err := cmd.Flags().GetBool("timestamp")
			if err != nil {
				return err
			}

			previous, err := cmd.Flags().GetBool("previous")
			if err != nil {
				return err
			}

			follow, err := cmd.Flags().GetBool("follow")
			if err != nil {
				return err
			}

			return tail(cmd.Context(), namespace, args, kubectl.PodLogOptions{
				Follow:     follow,
				Previous:   previous,
				Timestamps: timestamp,
				Since:      since,
			})
		},
	}

	cmd.MarkFlagRequired("namespace")
	cmd.Flags().StringP("namespace", "n", "", "kubectl namespace to use")
	cmd.Flags().StringP("since", "s", "", "a duration representing the logs since now you are interested in")
	cmd.Flags().BoolP("timestamp", "t", false, "include timestamps in logs")
	cmd.Flags().BoolP("previous", "p", false, "include logs of previous instances")
	cmd.Flags().BoolP("follow", "f", true, "follow the logs. defaults to true")

	return cmd
}

func tail(ctx context.Context, namespace string, labels []string, opts kubectl.PodLogOptions) error {
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

	set := makeSet()
	output := make(chan string)

	go func() {
		for pod := range pods {
			go func(pod *corev1.Pod) {
				if pod.Status.Phase == corev1.PodPending {
					err := wait.PollImmediate(time.Second, 20*time.Second, func() (done bool, err error) {
						p, err := ctl.GetPod(ctx, namespace, pod.Name)
						if err != nil {
							return false, err
						}
						return p.Status.Phase == corev1.PodRunning, nil
					})

					if err != nil {
						fmt.Printf("failed to poll pod: %v", err)
						return
					}
				}

				for _, container := range pod.Spec.Containers {

					key := strings.Join([]string{pod.Name, container.Name}, "/")
					if ok := set.add(key); !ok {
						continue
					}

					// copy opts
					opts := opts
					opts.Container = container.Name

					rc, err := ctl.StreamPodLogs(ctx, namespace, pod.Name, opts)
					if err != nil {
						fmt.Printf("failed to stream logs for pod \"%s:%s\": %v\n", pod.Name, container.Name, err)
						continue
					}

					prefix := pod.Name
					if len(pod.Spec.Containers) > 1 {
						prefix += "/" + container.Name
					}

					prefix = color.Color(prefix)

					go func(rc io.ReadCloser, key, prefix string) {
						defer rc.Close()
						defer set.remove(key)
						r := bufio.NewReader(rc)

						for {
							line, err := r.ReadString('\n')
							output <- fmt.Sprintf("%s  %s", prefix, line)
							if err != nil {
								fmt.Printf("%s  cannot continue reading: %v\n", prefix, err)
								return
							}
						}
					}(rc, key, prefix)
				}
			}(pod)
		}
	}()

	for out := range output {
		io.WriteString(os.Stdout, out)
	}

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
