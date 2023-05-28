package broadcast

import (
	"sync"
)

type messages struct {
	data []int
	seen map[int]struct{}
	lock *sync.RWMutex
}

func (m *messages) ReadAll() []int {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return append(make([]int, 0), m.data...)
}

func (m *messages) Add(msg int) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if _, ok := m.seen[msg]; ok {
		return
	}
	m.data = append(m.data, msg)
	m.seen[msg] = struct{}{}
}

func (m *messages) AddAll(msgs []int) {
	m.lock.Lock()
	defer m.lock.Unlock()
	for _, msg := range msgs {
		if _, ok := m.seen[msg]; ok {
			continue
		}
		m.data = append(m.data, msg)
		m.seen[msg] = struct{}{}
	}
}

func (m *messages) Remember(msg int) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.seen[msg] = struct{}{}
}

func (m *messages) RememberAll(msgs []int) {
	m.lock.Lock()
	defer m.lock.Unlock()
	for _, msg := range msgs {
		if _, ok := m.seen[msg]; ok {
			continue
		}
		m.seen[msg] = struct{}{}
	}
}

func (m *messages) ReadFiltered(msgs []int) []int {
	res := make([]int, 0, len(msgs))
	m.lock.RLock()
	defer m.lock.RUnlock()

	for _, msg := range msgs {
		if _, ok := m.seen[msg]; ok {
			continue
		}
		res = append(res, msg)
	}
	return res
}

func (m *messages) Contains(msg int) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()
	_, ok := m.seen[msg]
	return ok
}

func newMessages() *messages {
	return &messages{
		data: make([]int, 0, 1024),
		seen: make(map[int]struct{}, 1024),
		lock: &sync.RWMutex{},
	}
}

func newMessagesSeenOnly() *messages {
	return &messages{
		seen: make(map[int]struct{}, 1024),
		lock: &sync.RWMutex{},
	}
}

// type concurrentKnownMap struct {
// 	data  map[string]map[int]struct{}
// 	locks map[string]*sync.RWMutex
// 	lock  *sync.Mutex
// }
//
// func newKnownMap() *concurrentKnownMap {
// 	return &concurrentKnownMap{
// 		data:  make(map[string]map[int]struct{}),
// 		locks: make(map[string]*sync.RWMutex),
// 	}
// }
//
// func (cm *concurrentKnownMap) InitKey(key string) {
// 	_, ok := cm.locks[key]
// 	if !ok {
// 		cm.data[key] = make(map[int]struct{}, 1024)
// 		cm.locks[key] = &sync.RWMutex{}
// 	}
// }
//
// func (cm *concurrentKnownMap) Lock(key string) {
// 	lock, ok := cm.locks[key]
// 	if !ok {
// 		lock = &sync.RWMutex{}
// 		cm.locks[key] = lock
// 	}
// 	lock.Lock()
// }
//
// func (cm *concurrentKnownMap) Unlock(key string) {
// 	lock, ok := cm.locks[key]
// 	if !ok {
// 		return
// 	}
// 	lock.Unlock()
// }
//
// func (cm *concurrentKnownMap) RLock(key string) {
// 	lock, ok := cm.locks[key]
// 	if !ok {
// 		lock = &sync.RWMutex{}
// 		cm.locks[key] = lock
// 	}
// 	lock.RLock()
// }
//
// func (cm *concurrentKnownMap) RUnlock(key string) {
// 	cm.lock.Lock()
// 	defer cm.lock.Unlock()
// 	lock, ok := cm.locks[key]
// 	if !ok {
// 		return
// 	}
// 	lock.RUnlock()
// }
