package pages

import (
	"sync"
)

type DomainLocker struct {
	mutexes sync.Map
}

func NewDomainLocker() *DomainLocker {
	return &DomainLocker{
		mutexes: sync.Map{},
	}
}

func (m *DomainLocker) LockAny(any string) func() {
	value, _ := m.mutexes.LoadOrStore(any, &sync.Mutex{})
	mtx := value.(*sync.Mutex)
	mtx.Lock()

	return func() { mtx.Unlock() }
}
func (m *DomainLocker) Lock(domain *PageDomain) func() {
	return m.LockAny(domain.key())
}
