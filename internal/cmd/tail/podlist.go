package tail

import (
	"sync"
)

// set is a threadsafe set of strings
type set struct {
	set map[string]struct{}
	mu  sync.RWMutex
}

// add adds a value to the set. Returns true if a new value was added to the set.
func (p *set) add(value string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	_, ok := p.set[value]
	if !ok {
		p.set[value] = struct{}{}
	}
	return !ok
}

// remove a value from the set
func (p *set) remove(value string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.set, value)
}

// Length returns the how many values are in the set
func (p *set) Length() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.set)
}
