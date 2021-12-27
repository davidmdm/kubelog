package tail

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"sync"
	"time"

	"github.com/davidmdm/kubelog/internal/kubectl"
	"github.com/davidmdm/kubelog/internal/terminal"

	"github.com/davidmdm/kubelog/internal/cmd"

	"github.com/davidmdm/kubelog/internal/color"
	"github.com/spf13/cobra"
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

	cmd.Flags().StringP("namespace", "n", "", "kubectl namespace to use")
	cmd.Flags().StringP("since", "s", "", "a duration representing the logs since now you are interested in")
	cmd.Flags().BoolP("timestamp", "t", false, "include timestamps in logs")
	cmd.Flags().BoolP("previous", "p", false, "include logs of previous instances")
	cmd.Flags().BoolP("follow", "f", true, "follow the logs. defaults to true")

	return cmd
}

type container struct {
	Pod       string
	Container string
}

func tail(ctx context.Context, namespace string, labels []string, opts kubectl.PodLogOptions) error {
	ctl, err := kubectl.NewCtl(namespace)
	if err != nil {
		return fmt.Errorf("failed to connect to kubernetes: %w", err)
	}

	if namespace == "" {
		ns, err := cmd.SelectNamespace(ctx, ctl)
		if err != nil {
			return err
		}
		*ctl = ctl.WithNamespace(ns)
	}

	if len(labels) == 0 {
		labels = append(labels, "")
	}

	watchers := make([]<-chan kubectl.PodEvent, len(labels))
	for i, label := range labels {
		watcher, err := ctl.WatchPods(ctx, label)
		if err != nil {
			return fmt.Errorf("failed to watch pods: %w", err)
		}
		watchers[i] = watcher
	}

	containers := make(chan container)
	streams := MakeSyncMap[time.Time]()

	go func() {
		for podEvent := range JoinChannels(watchers...) {
			if podEvent.Pod == nil {
				continue
			}
			for _, c := range podEvent.Pod.Status.ContainerStatuses {
				key := path.Join(podEvent.Pod.Name, c.Name)
				running := c.State.Running
				if running == nil {
					continue
				}

				if value, loaded := streams.PutOrGet(key, running.StartedAt.Time); !loaded {
					containers <- container{Pod: podEvent.Pod.Name, Container: c.Name}
				} else if running.StartedAt.Time.After(value) {
					streams.Put(key, running.StartedAt.Time)
					containers <- container{Pod: podEvent.Pod.Name, Container: c.Name}
				}
			}
		}
		close(containers)
	}()

	output := make(chan string)
	done := make(chan struct{})
	wg := new(sync.WaitGroup)

	go func() {
		<-done
		wg.Wait()
		close(output)
	}()

	go func() {
		defer close(done)
		for c := range containers {
			wg.Add(1)
			go func(c container) {
				defer wg.Done()

				key := path.Join(c.Pod, c.Container)
				defer streams.Remove(key)

				// copy opts
				opts := opts
				opts.Container = c.Container

				rc, err := ctl.StreamPodLogs(ctx, c.Pod, opts)
				if err != nil {
					terminal.PrintErrf("failed to stream logs for pod \"%s:%s\": %v\n", c.Pod, c.Container, err)
					return
				}
				defer rc.Close()

				prefix := color.Color(key)
				r := bufio.NewReader(rc)

				for {
					line, err := r.ReadString('\n')
					if line != "" {
						output <- fmt.Sprintf("%s  %s", prefix, line)
					}
					if err != nil {
						if !errors.Is(err, context.Canceled) {
							terminal.PrintErrf("%s  cannot continue reading: %v\n", prefix, err)
						}
						return
					}
				}
			}(c)
		}
	}()

	for out := range output {
		io.WriteString(os.Stdout, out)
	}

	return ctx.Err()
}

func JoinChannels[T any](channels ...<-chan T) <-chan T {
	result := make(chan T)
	wg := new(sync.WaitGroup)

	for _, c := range channels {
		wg.Add(1)
		go func(c <-chan T) {
			defer wg.Done()
			for t := range c {
				result <- t
			}
		}(c)
	}

	go func() {
		wg.Wait()
		close(result)
	}()

	return result
}
