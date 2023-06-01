package consul

import (
	"math/rand"
	"sync"

	consul "github.com/hashicorp/consul/api"
)

type registryInterface interface {
	get(service string) *registryEntry
	expire(service string) // clears slice of instances
	remove(service string) // removes entry from registry
	addOrUpdate(service string, services []*consul.ServiceEntry)
}

type registry struct {
	entries *sync.Map
}

type registryEntry struct {
	services []*consul.ServiceEntry
	mu       sync.RWMutex
}

func (e *registryEntry) next() *consul.ServiceEntry {
	e.mu.Lock()
	defer e.mu.Unlock()

	if len(e.services) == 0 {
		return nil
	}
	//nolint:gosec
	return e.services[rand.Int()%len(e.services)]
}

func (r *registry) get(service string) *registryEntry {
	if result, ok := r.entries.Load(service); ok {
		return result.(*registryEntry)
	}

	return nil
}

func (r *registry) addOrUpdate(service string, services []*consul.ServiceEntry) {
	var entry *registryEntry

	// update
	if entry = r.get(service); entry != nil {
		entry.mu.Lock()
		defer entry.mu.Unlock()

		entry.services = services

		return
	}

	// add
	r.entries.Store(service, &registryEntry{
		services: services,
	})
}

func (r *registry) remove(service string) {
	r.entries.Delete(service)
}

func (r *registry) expire(service string) {
	var entry *registryEntry

	if entry = r.get(service); entry == nil {
		return
	}

	entry.mu.Lock()
	defer entry.mu.Unlock()

	entry.services = nil
}
