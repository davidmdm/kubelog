package tail

import (
	"sync"
)

// set is a threadsafe set of strings
type SyncMap[T any] struct {
	l *sync.Mutex
	m map[string]T
}

// PutOrGet will get the value from the map and if it fail to find it puts the given value to the map for the given key.
// the boolean returned reflects if the result was loaded from the map.
func (m SyncMap[T]) PutOrGet(key string, value T) (result T, loaded bool) {
	m.l.Lock()
	defer m.l.Unlock()
	if value, ok := m.m[key]; ok {
		return value, ok
	}
	m.m[key] = value
	return value, false
}

func (m SyncMap[T]) Put(key string, value T) {
	m.l.Lock()
	defer m.l.Unlock()
	m.m[key] = value
}

func (m SyncMap[T]) Remove(key string) {
	m.l.Lock()
	defer m.l.Unlock()
	delete(m.m, key)
}

func MakeSyncMap[T any]() SyncMap[T] {
	return SyncMap[T]{
		m: make(map[string]T),
		l: new(sync.Mutex),
	}
}
