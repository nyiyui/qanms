// Package goal provides a goal-based library for configuring WireGuard devices and routing.
// Note: only IPv4 supported yet.

package goal

type Machine struct {
	Interfaces   []Interface
	ForwardsIPv4 bool
	ForwardsIPv6 bool
}

type Interface struct {
	Name string

	PrivateKey Key

	// ListenPort is the device's listening port. set to -1 for nothing.
	ListenPort int

	Addresses []IPNet

	Peers []InterfacePeer

	// Broken is true if the interface a) has an address that is not assigned with ip, b) has a peer with an AllowedIPs that is not assigned with ip.
	Broken bool
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

type ApplierOptions struct {
	Linux ApplierOptionsLinux
}

type ApplierOptionsLinux struct {
	ReadWriteProc bool
}
