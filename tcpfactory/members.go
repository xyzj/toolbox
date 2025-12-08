package tcpfactory

import (
	"sync"

	"github.com/xyzj/toolbox/queue"
)

type members struct {
	locker       sync.RWMutex
	data         map[uint64]*tcpCore
	targets      map[string]uint64
	mulitTargets bool
}

// Store a new member
func (m *members) Store(sid uint64, value *tcpCore) {
	m.locker.Lock()
	m.data[sid] = value
	m.locker.Unlock()
}

// Len returns the number of members
func (m *members) Len() int {
	m.locker.RLock()
	l := len(m.data)
	m.locker.RUnlock()
	return l
}

// Shutdown a member by socket ID
func (m *members) Shutdown(sid uint64, reason string) {
	m.locker.Lock()
	if v, ok := m.data[sid]; ok {
		v.disconnect(reason)
		delete(m.data, sid)
		for k, id := range m.targets {
			if id == sid {
				delete(m.targets, k)
				if !m.mulitTargets {
					break
				}
			}
		}
	}
	m.locker.Unlock()
}

// ShutdownAll shuts down all members
func (m *members) ShutdownAll() {
	m.locker.Lock()
	defer m.locker.Unlock()
	for _, v := range m.data {
		v.disconnect("server shutdown")
	}
	m.targets = make(map[string]uint64)
	m.data = make(map[uint64]*tcpCore)
}

// SendTo sends a message to a specific member
func (m *members) SendTo(priority queue.Priority, target string, msgs ...*SendMessage) bool {
	m.locker.RLock()
	defer m.locker.RUnlock()
	if sockID, ok := m.targets[target]; ok {
		if v, ok := m.data[sockID]; ok {
			return v.writeTo(priority, target, msgs...)
		}
	}
	for _, v := range m.data {
		if v.writeTo(priority, target, msgs...) {
			m.targets[target] = v.sockID
			return true
		}
	}
	return false
}

// Delete removes a member by socket ID
func (m *members) Delete(sid uint64) {
	m.locker.Lock()
	delete(m.data, sid)
	for k, id := range m.targets {
		if id == sid {
			delete(m.targets, k)
			if !m.mulitTargets {
				break
			}
		}
	}
	m.locker.Unlock()
}

// NewMembers creates a new members instance
func newMembers(l int, mulitTargets bool) *members {
	return &members{
		locker:       sync.RWMutex{},
		data:         make(map[uint64]*tcpCore, l),
		targets:      make(map[string]uint64, l),
		mulitTargets: mulitTargets,
	}
}
