package spec

import (
	"bytes"
	"errors"
	"slices"

	"github.com/nyiyui/qrystal/goal"
)

type Spec struct {
	Networks []Network
}

func (s Spec) Clone() Spec {
	panic("not implemented yet")
	//newSpec := Spec{Networks: make([]Network, len(s.Networks))}
	//for i, sn := range s.Networks {
	//	newSpec.Networks[i] = sn.Clone()
	//}
	//return newSpec
}

func (s Spec) GetNetwork(name string) (n Network, ok bool) {
	i, ok := s.GetNetworkIndex(name)
	if !ok {
		return Network{}, false
	}
	return s.Networks[i], true
}

func (s Spec) GetNetworkIndex(name string) (i int, ok bool) {
	i = slices.IndexFunc(s.Networks, func(n Network) bool { return n.Name == name })
	if i == -1 {
		return 0, false
	}
	return i, true
}

type SpecCensored struct {
	Networks []NetworkCensored
}

func (a SpecCensored) Equal(b SpecCensored) bool {
	return slices.EqualFunc(a.Networks, b.Networks, func(a, b NetworkCensored) bool { return a.Equal(b) })
}

type Network struct {
	Name    string
	Devices []NetworkDevice
}

func (n Network) GetDevice(name string) (nd NetworkDevice, ok bool) {
	i, ok := n.GetDeviceIndex(name)
	if !ok {
		return NetworkDevice{}, false
	}
	return n.Devices[i], true
}

func (n Network) GetDeviceIndex(name string) (i int, ok bool) {
	i = slices.IndexFunc(n.Devices, func(nd NetworkDevice) bool { return nd.Name == name })
	return i, i != -1
}

type NetworkCensored struct {
	Name        string
	Devices     []NetworkDeviceCensored
	CensoredFor string
}

func (nc NetworkCensored) GetDevice(name string) (ndc NetworkDeviceCensored, ok bool) {
	i, ok := nc.GetDeviceIndex(name)
	if !ok {
		return NetworkDeviceCensored{}, false
	}
	return nc.Devices[i], true
}

func (nc NetworkCensored) GetDeviceIndex(name string) (i int, ok bool) {
	i = slices.IndexFunc(nc.Devices, func(ndc NetworkDeviceCensored) bool { return ndc.Name == name })
	return i, i != -1
}

func (a NetworkCensored) Equal(b NetworkCensored) bool {
	return a.Name == b.Name && slices.EqualFunc(a.Devices, b.Devices, func(a, b NetworkDeviceCensored) bool { return a.Equal(b) }) && a.CensoredFor == b.CensoredFor
}

func (n Network) CensorForDevice(censorFor string) NetworkCensored {
	nc := NetworkCensored{Name: n.Name, CensoredFor: censorFor}
	i := slices.IndexFunc(n.Devices, func(nd NetworkDevice) bool { return nd.Name == censorFor })
	if i == -1 {
		panic("censorFor device not in Network.Devices")
	}
	censorForDevice := n.Devices[i]
	for _, nd := range n.Devices {
		if censorForDevice.AccessAll || slices.Contains(censorForDevice.AccessOnly, nd.Name) {
			nc.Devices = append(nc.Devices, nd.NetworkDeviceCensored)
		}
	}
	return nc
}

type NetworkDevice struct {
	NetworkDeviceCensored

	AccessControl
}

type NetworkDeviceCensored struct {
	Name string
	// TODO: forwarding (keep this commented out for now)
	// // Route represents how a packet reaches this peer.
	// // Packets go through the peers listed (by name) in the slice.
	// // A nil slice means that no forwarding occurs.
	// // For example, {"A", "B"} means that a packet is forwarded to A to B to this peer.
	// Route []string
	// Endpoints is a unordered list of endpoints on which the peer is available on.
	Endpoints []string
	// EndpointChosen is whether the endpoint was chosen.
	// This value should always be false on the server.
	EndpointChosen bool
	// EndpointChosenIndex is the index of the chosen endpoint.
	// This value should always be zero on the server.
	EndpointChosenIndex int
	// Addresses is a list of IP networks that the peer represents.
	Addresses []goal.IPNet

	// ListenPort is the port that WireGuard will listen on.
	// This can be set by this peer.
	ListenPort int
	// This can be set by this peer.
	PublicKey goal.Key
	// This can be set by this peer.
	PresharedKey *goal.Key
	// PersistentKeepalive specifies how often a packet is sent to keep a connection alive.
	// Set to 0 to disable persistent keepalive.
	// This can be set by this peer.
	PersistentKeepalive goal.Duration
}

func ipNetEqual(a, b goal.IPNet) bool {
	return a.IP.Equal(b.IP) && bytes.Equal(a.Mask, b.Mask)
}

func (a NetworkDeviceCensored) Equal(b NetworkDeviceCensored) bool {
	return a.Name == b.Name && slices.Equal(a.Endpoints, b.Endpoints) && slices.EqualFunc(a.Addresses, b.Addresses, ipNetEqual) && a.ListenPort == b.ListenPort && a.PublicKey == b.PublicKey && (a.PresharedKey != nil && b.PresharedKey != nil && *a.PresharedKey == *b.PresharedKey || a.PresharedKey == nil && b.PresharedKey == nil) && a.PersistentKeepalive == b.PersistentKeepalive
}

type AccessControl struct {
	AccessAll  bool
	AccessOnly []string
}

func (a AccessControl) Validate() error {
	if a.AccessAll && a.AccessOnly != nil {
		return errors.New("AccessControl.AccessOnly is not nil and AccessControl.AccessAll is true")
	}
	return nil
}
