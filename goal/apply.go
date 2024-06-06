package goal

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"slices"

	"github.com/vishvananda/netlink"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func ApplyMachineDiff(a, b Machine, md MachineDiff, client *wgctrl.Client, handle *netlink.Handle) (err error) {
	for _, iface := range md.InterfacesRemoved {
		zap.S().Debugf("removing interface %s.", iface.Name)
		err = DeleteInterface(iface, client, handle)
		if err != nil {
			return
		}
	}
	for _, iface := range md.InterfacesAdded {
		zap.S().Debugf("adding interface %s.", iface.Name)
		err = CreateInterface(iface, client, handle)
		if err != nil {
			return
		}
	}
	for _, ifaceName := range md.InterfacesChanged {
		zap.S().Debugf("changing interface %s.", ifaceName)
		aIndex := slices.IndexFunc(a.Interfaces, func(iface Interface) bool { return iface.Name == ifaceName })
		bIndex := slices.IndexFunc(b.Interfaces, func(iface Interface) bool { return iface.Name == ifaceName })
		id := DiffInterface(&a.Interfaces[aIndex], &b.Interfaces[bIndex])
		err = ApplyInterfaceDiff(a.Interfaces[aIndex], b.Interfaces[bIndex], id, client, handle)
		if err != nil {
			return
		}
	}
	return nil
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
	zap.S().Debugf("adding link %s.", iface.Name)
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

	data, _ := json.MarshalIndent(id, "  ", "  ")
	zap.S().Debugf("interface diff:\n%s", data)

	zap.S().Debugf("configuring wg interface %s.", b.Name)
	// === configure wg interface ===
	peers := make([]wgtypes.PeerConfig, len(b.Peers))
	for i, peer := range b.Peers {
		zap.S().Debugf("resolving %s for peer %s.", peer.Endpoint, peer.Name)
		var endpoint *net.UDPAddr
		endpoint, err = net.ResolveUDPAddr("udp", peer.Endpoint)
		if err != nil {
			return err
		}
		peers[i] = wgtypes.PeerConfig{
			PublicKey:    wgtypes.Key(peer.PublicKey),
			PresharedKey: (*wgtypes.Key)(peer.PresharedKey),
			Endpoint:     endpoint,
			// TODO: PersistenKeepaliveInternal
			ReplaceAllowedIPs: true,
			AllowedIPs:        ipNetUtilToStd(peer.AllowedIPs),
		}
	}
	cfg := wgtypes.Config{
		PrivateKey:   (*wgtypes.Key)(&b.PrivateKey),
		ListenPort:   &b.ListenPort,
		ReplacePeers: true,
		Peers:        peers,
	}
	data, _ = json.MarshalIndent(cfg, "  ", "  ")
	zap.S().Debugf("wg interface configuration:\n%s", data)
	err = client.ConfigureDevice(b.Name, cfg)
	if err != nil {
		return fmt.Errorf("configuring wg interface: %w", err)
	}
	// CLEANUP: wg device is deleted when `ip link del` happens, so no cleanup is necessary.

	// === remove addresses to wg interface ===
	for _, change := range id.AddressesChanged {
		if change.Op != ChangeOpRemove {
			continue
		}
		zap.S().Debugf("removing address %s from wg interface", change.Value)
		addr := &netlink.Addr{
			IPNet: (*net.IPNet)(&change.Value),
		}
		err := handle.AddrDel(link, addr)
		if err != nil {
			return err
		}
	}
	// CLEANUP: re-add each address added
	defer func() {
		for _, change := range id.AddressesChanged {
			if change.Op != ChangeOpRemove {
				continue
			}
			addr := &netlink.Addr{
				IPNet: (*net.IPNet)(&change.Value),
			}
			err = handle.AddrAdd(link, addr)
			if err != nil {
				return
			}
		}
	}()

	// === add addresses to wg interface ===
	for _, change := range id.AddressesChanged {
		if change.Op != ChangeOpAdd {
			continue
		}
		zap.S().Debugf("adding address %s from wg interface", change.Value)
		addr := &netlink.Addr{
			IPNet: (*net.IPNet)(&change.Value),
		}
		err := handle.AddrAdd(link, addr)
		if err != nil {
			return err
		}
	}
	// CLEANUP: remove each address added
	defer func() {
		for _, change := range id.AddressesChanged {
			if change.Op != ChangeOpAdd {
				continue
			}
			addr := &netlink.Addr{
				IPNet: (*net.IPNet)(&change.Value),
			}
			err = handle.AddrDel(link, addr)
			if err != nil {
				return
			}
		}
	}()

	// === remove routes to wg interface ===
	for _, peer := range id.PeersRemoved {
		for _, allowedIP := range peer.AllowedIPs {
			zap.S().Debugf("removing route with dst %s from wg link", allowedIP)
			err = handle.RouteDel(&netlink.Route{
				LinkIndex: link.Attrs().Index,
				Dst:       (*net.IPNet)(&allowedIP),
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
					Dst:       (*net.IPNet)(&allowedIP),
					// TODO
				})
				if err != nil {
					return
				}
			}
		}
	}()

	// === add routes to wg interface ===
	for _, peer := range id.PeersRemoved {
		for _, allowedIP := range peer.AllowedIPs {
			zap.S().Debugf("adding route with dst %s from wg link", allowedIP)
			err = handle.RouteDel(&netlink.Route{
				LinkIndex: link.Attrs().Index,
				Dst:       (*net.IPNet)(&allowedIP),
				// TODO
			})
			if err != nil {
				return
			}
		}
	}
	// CLEANUP: remove each route added
	defer func() {
		for _, peer := range id.PeersRemoved {
			for _, allowedIP := range peer.AllowedIPs {
				err = handle.RouteAdd(&netlink.Route{
					LinkIndex: link.Attrs().Index,
					Dst:       (*net.IPNet)(&allowedIP),
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
