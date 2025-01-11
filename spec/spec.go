package spec

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"

	"github.com/nyiyui/qrystal/goal"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type Spec struct {
	Networks []Network
}

func (s Spec) Clone() Spec {
	newSpec := Spec{Networks: make([]Network, len(s.Networks))}
	for i, sn := range s.Networks {
		newSpec.Networks[i] = sn.Clone()
	}
	return newSpec
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

func (s SpecCensored) GetNetwork(name string) (n NetworkCensored, ok bool) {
	i, ok := s.GetNetworkIndex(name)
	if !ok {
		return NetworkCensored{}, false
	}
	return s.Networks[i], true
}

func (s SpecCensored) GetNetworkIndex(name string) (i int, ok bool) {
	i = slices.IndexFunc(s.Networks, func(n NetworkCensored) bool { return n.Name == name })
	if i == -1 {
		return 0, false
	}
	return i, true
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

func (n Network) Clone() Network {
	devices := make([]NetworkDevice, len(n.Devices))
	for i, nd := range n.Devices {
		devices[i] = nd.Clone()
	}
	return Network{n.Name, devices}
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

// GetForwardersFor returns a list of device names that can forward for the given device, and has a chosen forwarder and endpoint.
func (nc NetworkCensored) GetForwardersFor(name string) []string {
	forwarders := make([]string, 0)
	for _, ndc := range nc.Devices {
		if slices.Contains(ndc.ForwardsFor, name) && ndc.ForwarderAndEndpointChosen {
			forwarders = append(forwarders, ndc.Name)
		}
	}
	return forwarders
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

func (a NetworkDevice) Equal(b NetworkDevice) bool {
	return a.NetworkDeviceCensored.Equal(b.NetworkDeviceCensored) && a.AccessControl.Equal(b.AccessControl)
}

func (nd NetworkDevice) Clone() NetworkDevice {
	return NetworkDevice{
		NetworkDeviceCensored: nd.NetworkDeviceCensored.Clone(),
		AccessControl:         nd.AccessControl.Clone(),
	}
}

func (nd *NetworkDevice) UnmarshalJSON(data []byte) error {
	err := nd.NetworkDeviceCensored.UnmarshalJSON(data)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &nd.AccessControl)
	if err != nil {
		return err
	}
	return nil
}

type NetworkDeviceCensored struct {
	Name string
	// Endpoints is a unordered list of endpoints on which the peer is available on.
	Endpoints []string
	// EndpointChosen is whether the endpoint was chosen.
	// This value should always be false on the server.
	ForwarderAndEndpointChosen bool
	UsesForwarder              bool
	// EndpointChosenIndex is the index of the chosen endpoint.
	// This value should always be zero on the server.
	EndpointChosenIndex int
	// ForwarderChosenIndex is the name of the chosen forwarder.
	// This value should always be zero on the server.
	ForwarderChosenIndex int
	// Addresses is a list of IP networks that this peer represents.
	Addresses []goal.IPNet

	// ListenPort is the port that WireGuard will listen on.
	// Set to 0 to not specify a port.
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
	// ForwardsFor is the list of devices (in the same network) that this peer has access to, and can fowrard packets to.
	// Note that IPv6 forwarding is not supported yet.
	// This can be set by this peer.
	ForwardsFor []string
}

type networkDeviceCensoredJSON struct {
	Name                string
	Endpoints           []string
	Addresses           []goal.IPNet
	ListenPort          int
	PublicKey           goal.Key
	PresharedKey        *goal.Key
	PresharedKeyPath    string
	PersistentKeepalive goal.Duration
}

func (ndc *NetworkDeviceCensored) UnmarshalJSON(data []byte) error {
	var ndcj networkDeviceCensoredJSON
	err := json.Unmarshal(data, &ndcj)
	if err != nil {
		return err
	}
	if ndcj.PresharedKeyPath != "" {
		data, err := os.ReadFile(ndcj.PresharedKeyPath)
		if err != nil {
			return fmt.Errorf("reading PresharedKeyPath: %w", err)
		}
		presharedKey, err := wgtypes.ParseKey(string(data))
		if err != nil {
			return fmt.Errorf("parsing PresharedKeyPath: %w", err)
		}
		ndcj.PresharedKey = (*goal.Key)(&presharedKey)
	}
	ndc.Name = ndcj.Name
	ndc.Endpoints = ndcj.Endpoints
	ndc.Addresses = ndcj.Addresses
	ndc.ListenPort = ndcj.ListenPort
	ndc.PublicKey = ndcj.PublicKey
	ndc.PresharedKey = ndcj.PresharedKey
	ndc.PersistentKeepalive = ndcj.PersistentKeepalive
	return nil
}

func ipNetEqual(a, b goal.IPNet) bool {
	return a.IP.Equal(b.IP) && bytes.Equal(a.Mask, b.Mask)
}

func (a NetworkDeviceCensored) Equal(b NetworkDeviceCensored) bool {
	return a.Name == b.Name && slices.Equal(a.Endpoints, b.Endpoints) && slices.EqualFunc(a.Addresses, b.Addresses, ipNetEqual) && a.ListenPort == b.ListenPort && a.PublicKey == b.PublicKey && (a.PresharedKey != nil && b.PresharedKey != nil && *a.PresharedKey == *b.PresharedKey || a.PresharedKey == nil && b.PresharedKey == nil) && a.PersistentKeepalive == b.PersistentKeepalive && slices.Equal(a.ForwardsFor, b.ForwardsFor)
}

func (ndc NetworkDeviceCensored) Clone() NetworkDeviceCensored {
	ndc2 := NetworkDeviceCensored{
		Name:                       ndc.Name,
		ForwarderAndEndpointChosen: ndc.ForwarderAndEndpointChosen,
		EndpointChosenIndex:        ndc.EndpointChosenIndex,
		ForwarderChosenIndex:       ndc.ForwarderChosenIndex,
		ListenPort:                 ndc.ListenPort,
		PublicKey:                  ndc.PublicKey,
		PersistentKeepalive:        ndc.PersistentKeepalive,
	}
	ndc2.Endpoints = make([]string, len(ndc.Endpoints))
	copy(ndc2.Endpoints, ndc.Endpoints)
	ndc2.Addresses = make([]goal.IPNet, len(ndc.Addresses))
	for i, addr := range ndc.Addresses {
		ndc2.Addresses[i].IP = make([]byte, len(addr.IP))
		copy(ndc2.Addresses[i].IP, addr.IP)
		ndc2.Addresses[i].Mask = make([]byte, len(addr.Mask))
		copy(ndc2.Addresses[i].Mask, addr.Mask)
	}
	if ndc.PresharedKey != nil {
		presharedKey := new(goal.Key)
		copy(presharedKey[:], ndc.PresharedKey[:])
		ndc2.PresharedKey = presharedKey
	}
	return ndc2
}

type AccessControl struct {
	AccessAll  bool
	AccessOnly []string
}

func (a AccessControl) Equal(b AccessControl) bool {
	return a.AccessAll == b.AccessAll && slices.Equal(a.AccessOnly, b.AccessOnly)
}

func (a AccessControl) Validate() error {
	if a.AccessAll && a.AccessOnly != nil {
		return errors.New("AccessControl.AccessOnly is not nil and AccessControl.AccessAll is true")
	}
	return nil
}

func (a AccessControl) Clone() AccessControl {
	a2 := AccessControl{
		AccessAll: a.AccessAll,
	}
	if a.AccessOnly != nil {
		a2.AccessOnly = make([]string, len(a.AccessOnly))
		copy(a2.AccessOnly, a.AccessOnly)
	}
	return a2
}
