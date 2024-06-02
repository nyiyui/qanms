package goal

import (
	"bytes"
	"net"
	"sort"
	"strings"
	"time"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type Machine struct {
	Interfaces []Interface
}

type MachineDiff struct {
	InterfacesChanged []Change[Interface]
	InterfaceNoChange bool
}

func DiffMachine(a, b *Machine) MachineDiff {
	var diff MachineDiff

	sort.Slice(a.Interfaces, func(i, j int) bool { return a.Interfaces[i].Name < a.Interfaces[j].Name })
	sort.Slice(b.Interfaces, func(i, j int) bool { return b.Interfaces[i].Name < b.Interfaces[j].Name })

	var i, j int
	for i < len(a.Interfaces) && j < len(b.Interfaces) {
		cmp := strings.Compare(a.Interfaces[i].Name, b.Interfaces[j].Name)
		switch {
		case cmp < 0:
			diff.InterfacesChanged = append(diff.InterfacesChanged, Change[Interface]{Op: ChangeOpRemove, Value: a.Interfaces[i]})
			i++
		case cmp > 0:
			diff.InterfacesChanged = append(diff.InterfacesChanged, Change[Interface]{Op: ChangeOpAdd, Value: b.Interfaces[j]})
			j++
		default:
			interfaceDiff := DiffInterface(&a.Interfaces[i], &b.Interfaces[j])
			if interfaceDiff.NoChange() {
				diff.InterfacesChanged = append(diff.InterfacesChanged, Change[Interface]{Op: ChangeOpNoChange, Value: a.Interfaces[i]})
			}
			i++
			j++
		}
	}

	for ; i < len(a.Interfaces); i++ {
		diff.InterfacesChanged = append(diff.InterfacesChanged, Change[Interface]{Op: ChangeOpRemove, Value: a.Interfaces[i]})
	}

	for ; j < len(b.Interfaces); j++ {
		diff.InterfacesChanged = append(diff.InterfacesChanged, Change[Interface]{Op: ChangeOpAdd, Value: b.Interfaces[j]})
	}

	return diff
}

type Interface struct {
	Name string

	PrivateKey wgtypes.Key

	// ListenPort is the device's listening port. set to -1 for nothing.
	ListenPort int

	Peers []InterfacePeer
}

type InterfaceDiff struct {
	PrivateKeyChanged bool
	ListenPortChanged bool
	PeersChanged      []Change[InterfacePeer]
	PeersNoChange     bool
}

func (diff InterfaceDiff) NoChange() bool {
	return !diff.PrivateKeyChanged && !diff.ListenPortChanged && !diff.PeersNoChange
}

// DiffInterface returns a diff of Interface structs.
// a and b's Peers' order may change.
// Note that Peers are identified by its Peer.ID.
func DiffInterface(a, b *Interface) InterfaceDiff {
	var diff InterfaceDiff
	if !bytes.Equal(a.PrivateKey[:], b.PrivateKey[:]) {
		diff.PrivateKeyChanged = true
	}
	if a.ListenPort != b.ListenPort {
		diff.ListenPortChanged = true
	}
	sort.Slice(a.Peers, func(i, j int) bool { return bytes.Compare(a.Peers[i].ID[:], a.Peers[j].ID[:]) < 0 })
	sort.Slice(b.Peers, func(i, j int) bool { return bytes.Compare(b.Peers[i].ID[:], b.Peers[j].ID[:]) < 0 })

	diff.PeersNoChange = true
	var i, j int
	for i < len(a.Peers) && j < len(b.Peers) {
		cmp := bytes.Compare(a.Peers[i].ID[:], b.Peers[j].ID[:])
		switch {
		case cmp < 0:
			diff.PeersNoChange = false
			diff.PeersChanged = append(diff.PeersChanged, Change[InterfacePeer]{Op: ChangeOpRemove, Value: a.Peers[i]})
			i++
		case cmp > 0:
			diff.PeersNoChange = false
			diff.PeersChanged = append(diff.PeersChanged, Change[InterfacePeer]{Op: ChangeOpAdd, Value: b.Peers[j]})
			j++
		default:
			peerDiff := DiffInterfacePeer(&a.Peers[i], &b.Peers[j])
			if peerDiff.NoChange() {
				diff.PeersChanged = append(diff.PeersChanged, Change[InterfacePeer]{Op: ChangeOpNoChange, Value: a.Peers[i]})
			} else {
				diff.PeersNoChange = false
			}
			i++
			j++
		}
	}

	for ; i < len(a.Peers); i++ {
		diff.PeersNoChange = false
		diff.PeersChanged = append(diff.PeersChanged, Change[InterfacePeer]{Op: ChangeOpRemove, Value: a.Peers[i]})
	}

	for ; j < len(b.Peers); j++ {
		diff.PeersNoChange = false
		diff.PeersChanged = append(diff.PeersChanged, Change[InterfacePeer]{Op: ChangeOpAdd, Value: b.Peers[j]})
	}

	return diff
}

type InterfacePeer struct {
	ID [32]byte

	PublicKey wgtypes.Key

	PresharedKey wgtypes.Key

	// Endpoint as a string that will be looked up.
	// Set to an empty string for nothing.
	Endpoint string

	// PersistentKeepalive specifies how often a packet it sent to keep a connection alive.
	// Set to 0 to disable persistent keepalive.
	PersistentKeepalive time.Duration

	AllowedIPs []net.IPNet
}

type InterfacePeerDiff struct {
	PublicKeyChanged           bool
	PresharedKeyChanged        bool
	EndpointChanged            bool
	PersistentKeepaliveChanged bool
	AllowedIPsChanged          []Change[net.IPNet]
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
	if !bytes.Equal(a.PresharedKey[:], b.PresharedKey[:]) {
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
			diff.AllowedIPsChanged = append(diff.AllowedIPsChanged, Change[net.IPNet]{Op: ChangeOpRemove, Value: a.AllowedIPs[i]})
			i++
		case cmp > 0:
			diff.AllowedIPsNoChange = false
			diff.AllowedIPsChanged = append(diff.AllowedIPsChanged, Change[net.IPNet]{Op: ChangeOpAdd, Value: b.AllowedIPs[j]})
			j++
		default:
			if !bytes.Equal(a.AllowedIPs[i].Mask, b.AllowedIPs[j].Mask) {
				diff.AllowedIPsNoChange = false
				diff.AllowedIPsChanged = append(diff.AllowedIPsChanged, Change[net.IPNet]{Op: ChangeOpRemove, Value: a.AllowedIPs[i]})
				diff.AllowedIPsChanged = append(diff.AllowedIPsChanged, Change[net.IPNet]{Op: ChangeOpAdd, Value: b.AllowedIPs[j]})
			} else {
				diff.AllowedIPsChanged = append(diff.AllowedIPsChanged, Change[net.IPNet]{Op: ChangeOpNoChange, Value: a.AllowedIPs[i]})
			}
			i++
			j++
		}
	}

	for ; i < len(a.AllowedIPs); i++ {
		diff.AllowedIPsNoChange = false
		diff.AllowedIPsChanged = append(diff.AllowedIPsChanged, Change[net.IPNet]{Op: ChangeOpRemove, Value: a.AllowedIPs[i]})
	}

	for ; j < len(b.AllowedIPs); j++ {
		diff.AllowedIPsNoChange = false
		diff.AllowedIPsChanged = append(diff.AllowedIPsChanged, Change[net.IPNet]{Op: ChangeOpAdd, Value: b.AllowedIPs[j]})
	}
	return diff
}
