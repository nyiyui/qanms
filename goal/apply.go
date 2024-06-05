package goal

import (
	"errors"
	"fmt"
	"net"
	"slices"

	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func ApplyMachineDiff(a, b Machine, md MachineDiff, client *wgctrl.Client, handle *netlink.Handle) (err error) {
	for _, iface := range md.InterfacesRemoved {
		err = DeleteInterface(iface, client, handle)
		if err != nil {
			return
		}
	}
	for _, iface := range md.InterfacesAdded {
		err = CreateInterface(iface, client, handle)
		if err != nil {
			return
		}
	}
	for _, ifaceName := range md.InterfacesChanged {
		aIndex := slices.IndexFunc(a.Interfaces, func(iface Interface) bool { return iface.Name == ifaceName })
		bIndex := slices.IndexFunc(b.Interfaces, func(iface Interface) bool { return iface.Name == ifaceName })
		id := DiffInterface(&a.Interfaces[aIndex], &b.Interfaces[bIndex])
		err = ApplyInterfaceDiff(a.Interfaces[aIndex], b.Interfaces[bIndex], id, client, handle)
		if err != nil {
			return
		}
	}
	panic("not implemented yet")
}

func DeleteInterface(iface Interface, client *wgctrl.Client, handle *netlink.Handle) (err error) {
	// Steps:
	// - delete interface

	// === delete interface ===
	var link netlink.Link
	link, err = handle.LinkByName(iface.Name)
	if err != nil {
		return
	}
	err = handle.LinkDel(link)
	return
}

func CreateInterface(iface Interface, client *wgctrl.Client, handle *netlink.Handle) (err error) {
	// Steps:
	// - add interface

	if len(iface.Name) > 15 {
		return errors.New("interface name too long (max 15)")
	}

	// === add interface ===
	// emulates PR #464 (not landed in stable yet)
	// https://github.com/xaionaro-go/netlink/blob/fdd1f99835f135fb252d9e6fedd004c4b81601fd/link.go
	err = handle.LinkAdd(&netlink.GenericLink{
		LinkAttrs: netlink.LinkAttrs{
			Name: iface.Name,
		},
		LinkType: "wireguard",
	})
	if err != nil {
		return err
	}
	// CLEANUP: clean up created link
	defer func() {
		err = handle.LinkDel(&netlink.GenericLink{
			LinkAttrs: netlink.LinkAttrs{
				Name: iface.Name,
			},
			LinkType: "wireguard",
		})
	}()

	a := Interface{
		Name:       iface.Name,
		PrivateKey: iface.PrivateKey,
		ListenPort: iface.ListenPort,
	}
	id := DiffInterface(&a, &iface)
	return ApplyInterfaceDiff(a, iface, id, client, handle)
}

func ApplyInterfaceDiff(a, b Interface, id InterfaceDiff, client *wgctrl.Client, handle *netlink.Handle) (err error) {
	// Steps:
	// - configure wg interface
	// - remove addresses from wg interface
	// - add addresses to wg interface
	// - set MTU[^1]
	// - set DNS[^1]
	// - remove routes to wg interface
	// - add routes to wg interface
	// [^1]: not implemented, maybe in future work

	if a.Name != b.Name {
		panic("cannot apply diffs between names")
	}

	var link netlink.Link
	link, err = handle.LinkByName(b.Name)
	if err != nil {
		return err
	}

	// === configure wg interface ===
	peers := make([]wgtypes.PeerConfig, len(b.Peers))
	for i, peer := range b.Peers {
		var endpoint *net.UDPAddr
		endpoint, err = net.ResolveUDPAddr("udp", peer.Endpoint)
		if err != nil {
			return
		}
		peers[i] = wgtypes.PeerConfig{
			PublicKey:    peer.PublicKey,
			PresharedKey: &peer.PresharedKey,
			Endpoint:     endpoint,
			AllowedIPs:   peer.AllowedIPs,
		}
	}
	err = client.ConfigureDevice(b.Name, wgtypes.Config{
		PrivateKey:   &b.PrivateKey,
		ListenPort:   &b.ListenPort,
		ReplacePeers: true,
		Peers:        peers,
	})
	if err != nil {
		return err
	}
	// CLEANUP: wg device is deleted when `ip link del` happens, so no cleanup is necessary.

	// === add addresses to wg interface ===
	for _, peer := range id.PeersRemoved {
		for j, allowedIP := range peer.AllowedIPs {
			addr := &netlink.Addr{
				IPNet: &allowedIP,
				Label: fmt.Sprintf("%s-%d", peer.Name, j),
			}
			err := handle.AddrAdd(link, addr)
			if err != nil {
				return err
			}
		}
	}
	// CLEANUP: remove each address added from device
	defer func() {
		for _, peer := range id.PeersRemoved {
			for j, allowedIP := range peer.AllowedIPs {
				addr := &netlink.Addr{
					IPNet: &allowedIP,
					Label: fmt.Sprintf("%s-%d", peer.Name, j),
				}
				err = handle.AddrDel(link, addr)
				if err != nil {
					return
				}
			}
		}
	}()

	// === add addresses to wg interface ===
	for _, peer := range id.PeersAdded {
		for j, allowedIP := range peer.AllowedIPs {
			addr := &netlink.Addr{
				IPNet: &allowedIP,
				Label: fmt.Sprintf("%s-%d", peer.Name, j),
			}
			err := handle.AddrAdd(link, addr)
			if err != nil {
				return err
			}
		}
	}
	// CLEANUP: delete each address added to device
	defer func() {
		for _, peer := range id.PeersAdded {
			for j, allowedIP := range peer.AllowedIPs {
				addr := &netlink.Addr{
					IPNet: &allowedIP,
					Label: fmt.Sprintf("%s-%d", peer.Name, j),
				}
				err = handle.AddrDel(link, addr)
				if err != nil {
					return
				}
			}
		}
	}()

	// === remove routes to wg interface ===
	for _, peer := range id.PeersRemoved {
		for _, allowedIP := range peer.AllowedIPs {
			err = handle.RouteDel(&netlink.Route{
				LinkIndex: link.Attrs().Index,
				Dst:       &allowedIP,
				// TODO
			})
			if err != nil {
				return
			}
		}
	}
	// CLEANUP: re-add each route removed
	defer func() {
		for _, peer := range id.PeersRemoved {
			for _, allowedIP := range peer.AllowedIPs {
				err = handle.RouteAdd(&netlink.Route{
					LinkIndex: link.Attrs().Index,
					Dst:       &allowedIP,
					// TODO
				})
				if err != nil {
					return
				}
			}
		}
	}()

	// === remove routes to wg interface ===
	for _, peer := range id.PeersRemoved {
		for _, allowedIP := range peer.AllowedIPs {
			err = handle.RouteDel(&netlink.Route{
				LinkIndex: link.Attrs().Index,
				Dst:       &allowedIP,
				// TODO
			})
			if err != nil {
				return
			}
		}
	}
	// CLEANUP: re-add each route removed
	defer func() {
		for _, peer := range id.PeersRemoved {
			for _, allowedIP := range peer.AllowedIPs {
				err = handle.RouteAdd(&netlink.Route{
					LinkIndex: link.Attrs().Index,
					Dst:       &allowedIP,
					// TODO
				})
				if err != nil {
					return
				}
			}
		}
	}()
	return
}
