package kubectl

import (
	"sync"

	"github.com/davidmdm/kubelog/internal/util"
)

// podList list is a threadsafe list of pods
type podList struct {
	pods []string
	mu   sync.RWMutex
}

// add adds pod to the podlist in a thread safe manner. Returns true if pod was added.
func (p *podList) add(pod string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !util.HasString(p.pods, pod) {
		p.pods = append(p.pods, pod)
		return true
	}
	return false
}

// remove removes a pod from the podlist in a thread safe manner. Returns true if pod was found and removed.
func (p *podList) remove(pod string) bool {
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

// has checks if a pod is contained within the podlist
func (p *podList) has(pod string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return util.HasString(p.pods, pod)
}

// Length returns the how many pods are in the list
func (p *podList) Length() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.pods)
}
