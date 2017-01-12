package main

import (
	"net"
	"sync"
)

type InterfaceList struct {
	mu         sync.RWMutex
	interfaces []net.Interface
}

func NewInterfaceList() *InterfaceList {
	return &InterfaceList{interfaces: make([]net.Interface, 0)}
}

func (il *InterfaceList) ClearAndAppend(interfaces []net.Interface) {
	il.mu.Lock()
	defer il.mu.Unlock()

	il.interfaces = make([]net.Interface, 0)

	for _, iface := range interfaces {
		il.interfaces = append(il.interfaces, iface)
	}
}

func (il *InterfaceList) Append(iface net.Interface) {
	il.mu.Lock()
	defer il.mu.Unlock()

	il.interfaces = append(il.interfaces, iface)
}

func (il *InterfaceList) Get(i int) net.Interface {
	il.mu.RLock()
	defer il.mu.RUnlock()

	return il.interfaces[i]
}

func (il *InterfaceList) All() []net.Interface {
	il.mu.RLock()
	defer il.mu.RUnlock()

	interfaces := make([]net.Interface, len(il.interfaces))

	for i, iface := range il.interfaces {
		interfaces[i] = iface
	}

	return interfaces
}

func (il *InterfaceList) Len() int {
	il.mu.RLock()
	defer il.mu.RUnlock()

	return len(il.interfaces)
}
