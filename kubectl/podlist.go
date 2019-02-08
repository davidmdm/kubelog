package kubectl

import (
	"sync"

	"github.com/davidmdm/kubelog/util"
)

// PodList list is a threadsafe list of pods
type PodList struct {
	pods []string
	mu   sync.Mutex
}

// Add adds pod to the podlist in a thread safe manner. Returns true if pod was added.
func (p *PodList) Add(pod string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !util.HasString(p.pods, pod) {
		p.pods = append(p.pods, pod)
		return true
	}
	return false
}

// Remove removes a pod from the podlist in a thread safe manner. Returns true if pod was found and removed.
func (p *PodList) Remove(pod string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	for i := range p.pods {
		if p.pods[i] == pod {
			p.pods = append(p.pods[:i], p.pods[i+1:]...)
			return true
		}
	}
	return false
}

// Has checks if a pod is contained within the podlist
func (p *PodList) Has(pod string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return util.HasString(p.pods, pod)
}

// Length returns the how many pods are in the list
func (p *PodList) Length() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.pods)
}
