package kubectl

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type K8Ctl struct {
	clientSet *kubernetes.Clientset
}

func NewCtl() (*K8Ctl, error) {
	cfg, err := clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
	if err != nil {
		return nil, fmt.Errorf("failed to construct k8 config: %w", err)
	}

	clientSet, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate k8 clientset: %w", err)
	}

	return &K8Ctl{clientSet}, nil
}

func (ctl K8Ctl) GetNamespaces(ctx context.Context) ([]corev1.Namespace, error) {
	var namespaces []corev1.Namespace
	var continueToken string

	for {
		namespaceList, err := ctl.clientSet.
			CoreV1().
			Namespaces().
			List(ctx, v1.ListOptions{Continue: continueToken})

		if err != nil {
			return nil, fmt.Errorf("failed to list namespaces: %w", err)
		}

		namespaces = append(namespaces, namespaceList.Items...)

		if continueToken = namespaceList.Continue; continueToken == "" {
			break
		}
	}

	return namespaces, nil
}

func (ctl K8Ctl) GetPods(ctx context.Context, namespace string, labelSelector string) ([]corev1.Pod, error) {
	var continueToken string
	var pods []corev1.Pod

	for {
		podList, err := ctl.clientSet.
			CoreV1().
			Pods(namespace).
			List(ctx, v1.ListOptions{
				LabelSelector: labelSelector,
				Continue:      continueToken,
			})

		if err != nil {
			return pods, fmt.Errorf("failed to list pods: %w", err)
		}

		pods = append(pods, podList.Items...)

		if continueToken = podList.Continue; continueToken == "" {
			break
		}
	}

	return pods, nil
}

type PodLogOptions struct {
	Container  string
	Follow     bool
	Previous   bool
	Timestamps bool
	Since      *time.Duration
}

func (ctl K8Ctl) StreamPodLogs(ctx context.Context, namespace string, name string, opts PodLogOptions) (io.ReadCloser, error) {
	return ctl.clientSet.
		CoreV1().
		Pods(namespace).
		GetLogs(name, &corev1.PodLogOptions{
			Container:  opts.Container,
			Follow:     opts.Follow,
			Previous:   opts.Previous,
			Timestamps: opts.Timestamps,

			SinceSeconds: func() *int64 {
				if opts.Since == nil {
					return nil
				}
				seconds := int64(opts.Since.Seconds())
				return &seconds
			}(),
		}).
		Stream(ctx)
}

type PodEvent struct {
	Type watch.EventType
	Pod  *corev1.Pod
}

func (ctl K8Ctl) WatchPods(ctx context.Context, namespace string, labelSelector string) (<-chan PodEvent, error) {
	watcher, err := ctl.clientSet.
		CoreV1().
		Pods(namespace).
		Watch(ctx, v1.ListOptions{LabelSelector: labelSelector})

	if err != nil {
		return nil, fmt.Errorf("failed to watch pods: %w", err)
	}

	podEvents := make(chan PodEvent)
	events := watcher.ResultChan()

	go func() {
		for {
			select {
			case <-ctx.Done():
				watcher.Stop()
				return
			case evt := <-events:
				pod, ok := evt.Object.(*corev1.Pod)
				if !ok {
					continue
				}
				podEvents <- PodEvent{Type: evt.Type, Pod: pod}
			}
		}
	}()

	return podEvents, nil
}
