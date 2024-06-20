// Package goal provides a goal-based library for configuring WireGuard devices and routing.
// Note: only IPv4 supported yet.

package goal

import (
	"bytes"
	"sort"
)

type Machine struct {
	Interfaces []Interface
}

type MachineDiff struct {
	InterfacesAdded   []Interface
	InterfacesRemoved []Interface
	InterfacesChanged []string
	InterfaceNoChange bool
}

func DiffMachine(a, b *Machine) MachineDiff {
	var diff MachineDiff

	aNames := map[string]int{}
	for i, peer := range a.Interfaces {
		aNames[peer.Name] = i
	}
	bNames := map[string]int{}
	for i, peer := range b.Interfaces {
		bNames[peer.Name] = i
	}
	for _, peer := range a.Interfaces {
		if _, ok := bNames[peer.Name]; !ok {
			diff.InterfacesRemoved = append(diff.InterfacesRemoved, peer)
		}
	}
	for _, peer := range b.Interfaces {
		if _, ok := aNames[peer.Name]; !ok {
			diff.InterfacesAdded = append(diff.InterfacesAdded, peer)
		}
	}
	for i, peer := range a.Interfaces {
		_, ok1 := aNames[peer.Name]
		j, ok2 := bNames[peer.Name]
		if !ok1 || !ok2 {
			continue
		}
		d := DiffInterface(&a.Interfaces[i], &b.Interfaces[j])
		if !d.NoChange() {
			diff.InterfacesChanged = append(diff.InterfacesChanged, b.Interfaces[j].Name)
		}
	}

	return diff
}

type Interface struct {
	Name string

	PrivateKey Key

	// ListenPort is the device's listening port. set to -1 for nothing.
	ListenPort int

	Addresses []IPNet

	Peers []InterfacePeer
	// TODO: forwarding
}

type InterfaceDiff struct {
	PrivateKeyChanged bool
	ListenPortChanged bool
	AddressesAdded    []IPNet
	AddressesRemoved  []IPNet
	AddressesNoChange bool
	PeersAdded        []InterfacePeer
	PeersRemoved      []InterfacePeer
	PeersChanged      []string
}

func (diff InterfaceDiff) NoChange() bool {
	return !diff.PrivateKeyChanged && !diff.ListenPortChanged && diff.AddressesNoChange && len(diff.PeersAdded) == 0 && len(diff.PeersRemoved) == 0 && len(diff.PeersChanged) == 0
}

// DiffInterface returns a diff of Interface structs.
// a and b's Peers' order may change.
// Note that Peers are identified by its Peer.ID.
func DiffInterface(a, b *Interface) InterfaceDiff {
	if a.Name != b.Name {
		panic("cannot diff between names")
	}
	var diff InterfaceDiff
	if !bytes.Equal(a.PrivateKey[:], b.PrivateKey[:]) {
		diff.PrivateKeyChanged = true
	}
	if a.ListenPort != b.ListenPort {
		diff.ListenPortChanged = true
	}
	aNames := map[string]int{}
	for i, peer := range a.Peers {
		aNames[peer.Name] = i
	}
	bNames := map[string]int{}
	for i, peer := range b.Peers {
		bNames[peer.Name] = i
	}
	for _, peer := range a.Peers {
		if _, ok := bNames[peer.Name]; !ok {
			diff.PeersRemoved = append(diff.PeersRemoved, peer)
		}
	}
	for _, peer := range b.Peers {
		if _, ok := aNames[peer.Name]; !ok {
			diff.PeersAdded = append(diff.PeersAdded, peer)
		}
	}
	for i, peer := range a.Peers {
		_, ok1 := aNames[peer.Name]
		j, ok2 := bNames[peer.Name]
		if !ok1 || !ok2 {
			continue
		}
		d := DiffInterfacePeer(&a.Peers[i], &b.Peers[j])
		if !d.NoChange() {
			diff.PeersChanged = append(diff.PeersChanged, b.Peers[j].Name)
		}
	}
	sort.Slice(a.Addresses, func(i, j int) bool {
		return bytes.Compare(a.Addresses[i].IP[:], a.Addresses[j].IP[:]) < 0
	})
	sort.Slice(b.Addresses, func(i, j int) bool {
		return bytes.Compare(b.Addresses[i].IP[:], b.Addresses[j].IP[:]) < 0
	})
	diff.AddressesNoChange = true
	var i, j int
	for i < len(a.Addresses) && j < len(b.Addresses) {
		cmp := bytes.Compare(a.Addresses[i].IP, b.Addresses[j].IP)
		switch {
		case cmp < 0:
			diff.AddressesNoChange = false
			diff.AddressesRemoved = append(diff.AddressesRemoved, a.Addresses[i])
			i++
		case cmp > 0:
			diff.AddressesNoChange = false
			diff.AddressesAdded = append(diff.AddressesAdded, a.Addresses[i])
			j++
		default:
			if !bytes.Equal(a.Addresses[i].Mask, b.Addresses[j].Mask) {
				diff.AddressesNoChange = false
				diff.AddressesRemoved = append(diff.AddressesRemoved, a.Addresses[i])
				diff.AddressesAdded = append(diff.AddressesAdded, b.Addresses[i])
			}
			i++
			j++
		}
	}

	for ; i < len(a.Addresses); i++ {
		diff.AddressesNoChange = false
		diff.AddressesRemoved = append(diff.AddressesRemoved, a.Addresses[i])
	}

	for ; j < len(b.Addresses); j++ {
		diff.AddressesNoChange = false
		diff.AddressesAdded = append(diff.AddressesAdded, b.Addresses[j])
	}
	return diff
}

type InterfacePeer struct {
	Name string

	PublicKey Key

	PresharedKey *Key

	// Endpoint as a string that will be looked up.
	// Set to an empty string for nothing.
	Endpoint string

	// PersistentKeepalive specifies how often a packet is sent to keep a connection alive.
	// Set to 0 to disable persistent keepalive.
	PersistentKeepalive Duration

	AllowedIPs []IPNet
}

type InterfacePeerDiff struct {
	PublicKeyChanged           bool
	PresharedKeyChanged        bool
	EndpointChanged            bool
	PersistentKeepaliveChanged bool
	AllowedIPsChanged          []Change[IPNet]
	AllowedIPsNoChange         bool
}

func (diff InterfacePeerDiff) NoChange() bool {
	if diff.PublicKeyChanged || diff.PresharedKeyChanged || diff.EndpointChanged || diff.PersistentKeepaliveChanged {
		return false
	}
	return diff.AllowedIPsNoChange
}

type Change[T any] struct {
	Op    ChangeOp
	Value T
}

type ChangeOp int

const (
	ChangeOpNoChange = iota
	ChangeOpAdd
	ChangeOpRemove
)

// DiffInterfacePeer returns a diff of InterfacePeer structs.
// a and b's Endpoints' order may change.
// Note that Endpoints are diffed by the string, and the addresses are not resolved.
func DiffInterfacePeer(a, b *InterfacePeer) InterfacePeerDiff {
	var diff InterfacePeerDiff
	if !bytes.Equal(a.PublicKey[:], b.PublicKey[:]) {
		diff.PublicKeyChanged = true
	}
	if (a.PresharedKey == nil) != (b.PresharedKey == nil) || ((a.PresharedKey != nil && b.PresharedKey != nil) && !bytes.Equal(a.PresharedKey[:], b.PresharedKey[:])) {
		diff.PresharedKeyChanged = true
	}
	if a.Endpoint != b.Endpoint {
		diff.EndpointChanged = true
	}
	if a.PersistentKeepalive != b.PersistentKeepalive {
		diff.PersistentKeepaliveChanged = true
	}
	sort.Slice(a.AllowedIPs, func(i, j int) bool {
		return bytes.Compare(a.AllowedIPs[i].IP[:], a.AllowedIPs[j].IP[:]) < 0
	})
	sort.Slice(b.AllowedIPs, func(i, j int) bool {
		return bytes.Compare(b.AllowedIPs[i].IP[:], b.AllowedIPs[j].IP[:]) < 0
	})

	diff.AllowedIPsNoChange = true
	var i, j int
	for i < len(a.AllowedIPs) && j < len(b.AllowedIPs) {
		cmp := bytes.Compare(a.AllowedIPs[i].IP, b.AllowedIPs[j].IP)
		switch {
		case cmp < 0:
			diff.AllowedIPsNoChange = false
			diff.AllowedIPsChanged = append(diff.AllowedIPsChanged, Change[IPNet]{Op: ChangeOpRemove, Value: a.AllowedIPs[i]})
			i++
		case cmp > 0:
			diff.AllowedIPsNoChange = false
			diff.AllowedIPsChanged = append(diff.AllowedIPsChanged, Change[IPNet]{Op: ChangeOpAdd, Value: b.AllowedIPs[j]})
			j++
		default:
			if !bytes.Equal(a.AllowedIPs[i].Mask, b.AllowedIPs[j].Mask) {
				diff.AllowedIPsNoChange = false
				diff.AllowedIPsChanged = append(diff.AllowedIPsChanged, Change[IPNet]{Op: ChangeOpRemove, Value: a.AllowedIPs[i]})
				diff.AllowedIPsChanged = append(diff.AllowedIPsChanged, Change[IPNet]{Op: ChangeOpAdd, Value: b.AllowedIPs[j]})
			} else {
				diff.AllowedIPsChanged = append(diff.AllowedIPsChanged, Change[IPNet]{Op: ChangeOpNoChange, Value: a.AllowedIPs[i]})
			}
			i++
			j++
		}
	}

	for ; i < len(a.AllowedIPs); i++ {
		diff.AllowedIPsNoChange = false
		diff.AllowedIPsChanged = append(diff.AllowedIPsChanged, Change[IPNet]{Op: ChangeOpRemove, Value: a.AllowedIPs[i]})
	}

	for ; j < len(b.AllowedIPs); j++ {
		diff.AllowedIPsNoChange = false
		diff.AllowedIPsChanged = append(diff.AllowedIPsChanged, Change[IPNet]{Op: ChangeOpAdd, Value: b.AllowedIPs[j]})
	}
	return diff
}
