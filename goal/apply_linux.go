//go:build linux

package goal

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/vishvananda/netlink"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type Applier struct {
	client            *wgctrl.Client
	handle            *netlink.Handle
	managedInterfaces []string
	readWriteProc     bool
}

func NewApplier(opt ApplierOptions) (*Applier, error) {
	return NewApplierLinux(nil, nil, nil, opt.Linux.ReadWriteProc)
}

func NewApplierLinux(client *wgctrl.Client, handle *netlink.Handle, managedInterfaces []string, readWriteProc bool) (*Applier, error) {
	var err error
	if client == nil {
		client, err = wgctrl.New()
		if err != nil {
			return nil, fmt.Errorf("creating wgctrl client: %w", err)
		}
	}
	if handle == nil {
		handle, err = netlink.NewHandle()
		if err != nil {
			return nil, fmt.Errorf("creating netlink handle: %w", err)
		}
	}
	return &Applier{
		client:            client,
		handle:            handle,
		managedInterfaces: managedInterfaces,
		readWriteProc:     readWriteProc,
	}, nil
}

// ApplyMachine applies the given Machine to the system.
// If any error is encountered during application, this function immediately bails out, potentially leaving the system in an inconsistent state.
// (Assuming the error is temporary, you can just rerun this function to get to your goal state.)
func (a *Applier) ApplyMachine(m Machine) (err error) {
	devices, err := a.client.Devices()
	if err != nil {
		return fmt.Errorf("getting wg devices: %w", err)
	}
	deviceNames := make([]string, len(devices))
	for i, device := range devices {
		deviceNames[i] = device.Name
	}
	machineInterfaces := make([]string, len(m.Interfaces))
	for i, iface := range m.Interfaces {
		machineInterfaces[i] = iface.Name
	}
	less := func(x, y string) bool { return x < y }
	if unmanaged := setDifference(machineInterfaces, a.managedInterfaces, less); len(unmanaged) != 0 {
		// trying to manage an interface that is not managed
		// edit managedInterfaces to include the unmanaged interfaces
		a.managedInterfaces = append(a.managedInterfaces, unmanaged...)
	}
	deletedInterfaces := setIntersection(setDifference(a.managedInterfaces, machineInterfaces, less), deviceNames, less)
	createdInterfaces := setDifference(machineInterfaces, deviceNames, less)

	for _, ifaceName := range deletedInterfaces {
		zap.S().Debugf("removing interface %s…", ifaceName)
		i := slices.IndexFunc(m.Interfaces, func(iface Interface) bool { return iface.Name == ifaceName })
		err = deleteInterface(m.Interfaces[i], a.handle)
		if err != nil {
			return fmt.Errorf("removing interface %s: %w", ifaceName, err)
		}
	}

	for _, ifaceName := range createdInterfaces {
		zap.S().Debugf("adding interface %s…", ifaceName)
		i := slices.IndexFunc(m.Interfaces, func(iface Interface) bool { return iface.Name == ifaceName })
		err = createInterface(m.Interfaces[i], a.client, a.handle)
		if err != nil {
			return fmt.Errorf("adding interface %s: %w", ifaceName, err)
		}
	}

	updatedInterfaces := setDifference(a.managedInterfaces, deletedInterfaces, less)
	for _, ifaceName := range updatedInterfaces {
		zap.S().Debugf("updating interface %s…", ifaceName)
		i := slices.IndexFunc(m.Interfaces, func(iface Interface) bool { return iface.Name == ifaceName })
		err = a.updateInterface(m.Interfaces[i])
		if err != nil {
			return fmt.Errorf("updating interface %s: %w", ifaceName, err)
		}
	}
	if a.readWriteProc {
		data, err := os.ReadFile("/proc/sys/net/ipv4/ip_forward")
		if err != nil {
			return fmt.Errorf("reading /proc/sys/net/ipv4/ip_forward: %w", err)
		}
		value, err := strconv.Atoi(string(bytes.TrimSpace(data)))
		if err != nil {
			return fmt.Errorf("parsing /proc/sys/net/ipv4/ip_forward: %w", err)
		}
		oldValue := value == 1
		if oldValue != m.ForwardsIPv4 {
			var data []byte
			if m.ForwardsIPv4 {
				data = []byte("1")
			} else {
				data = []byte("0")
			}
			err := os.WriteFile("/proc/sys/net/ipv4/ip_forward", data, 0444)
			if err != nil {
				return fmt.Errorf("writing /proc/sys/net/ipv4/ip_forward: %w", err)
			}
		}
	}
	if m.ForwardsIPv6 {
		panic("not implemented yet")
	}
	return nil
}

func deleteInterface(iface Interface, handle *netlink.Handle) (err error) {
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

func createInterface(iface Interface, client *wgctrl.Client, handle *netlink.Handle) (err error) {
	// Steps:
	// - add interface

	if len(iface.Name) > 15 {
		return errors.New("interface name too long (max 15)")
	}

	// === add interface ===
	// emulates PR #464 (not landed in stable yet)
	// https://github.com/xaionaro-go/netlink/blob/fdd1f99835f135fb252d9e6fedd004c4b81601fd/link.go
	// ip link add dev <iface.Name> type wireguard
	zap.S().Debugf("adding link %s.", iface.Name)
	err = handle.LinkAdd(&netlink.GenericLink{
		LinkAttrs: netlink.LinkAttrs{
			Name: iface.Name,
		},
		LinkType: "wireguard",
	})
	if err != nil {
		return fmt.Errorf("adding link %s: %w", iface.Name, err)
	}
	return
}

func (a *Applier) updateInterface(iface Interface) error {
	// Steps:
	// - configure wg interface
	// - remove/add addresses from wg interface
	// - set MTU[^1]
	// - set DNS[^1]
	// - set up link (if applicable)
	// - remove/add routes to wg interface
	// [^1]: not implemented, maybe in future work

	var err error
	var link netlink.Link
	link, err = a.handle.LinkByName(iface.Name)
	if err != nil {
		// should not return with a "link not found," since that would fail on the wgctrl.Client.Device call
		return err
	}

	// === configure wg interface ===
	err = a.configureWireguard(iface)
	if err != nil {
		return err
	}

	// === remove/add addresses from wg interface ===
	err = a.updateInterfaceAddresses(iface, link)
	if err != nil {
		return err
	}

	// === set up link ===
	err = netlink.LinkSetUp(link)
	if err != nil {
		return fmt.Errorf("link set up: %w", err)
	}

	// === remove/add routes to wg interface ===
	err = a.applyInterfaceRoutes(iface, link)
	if err != nil {
		return err
	}
	zap.S().Debug("applied interface diff.")
	return nil
}

func (a *Applier) updateInterfaceAddresses(iface Interface, link netlink.Link) error {
	var tasks []addressTask
	linkAddrs_, err := a.handle.AddrList(link, netlink.FAMILY_ALL)
	if err != nil {
		return fmt.Errorf("listing addresses on wg interface: %w", err)
	}
	linkAddrs := make([]IPNet, len(linkAddrs_))
	for i, addr := range linkAddrs_ {
		linkAddrs[i] = IPNet(*addr.IPNet)
	}
	removedIPs := setDifference(linkAddrs, iface.Addresses, lessIPNet)
	for _, ip := range removedIPs {
		tasks = append(tasks, addressTask{add: false, ip: ip})
	}
	addedIPs := setDifference(iface.Addresses, linkAddrs, lessIPNet)
	for _, ip := range addedIPs {
		tasks = append(tasks, addressTask{add: true, ip: ip})
	}
	tasksStrings := make([]string, len(tasks))
	for i, task := range tasks {
		tasksStrings[i] = task.String()
	}
	zap.S().Debugf("changing %d addresses to wg interface:\n%s", len(tasks), strings.Join(tasksStrings, "\n"))
	for _, task := range tasks {
		if task.add {
			err = a.handle.AddrAdd(link, &netlink.Addr{
				IPNet: (*net.IPNet)(&task.ip),
			})
		} else {
			err = a.handle.AddrDel(link, &netlink.Addr{
				IPNet: (*net.IPNet)(&task.ip),
			})
		}
		if err != nil {
			return fmt.Errorf("task to wg interface (%s) failed: %w", task, err)
		}
	}
	return nil
}

func (a *Applier) applyInterfaceRoutes(iface Interface, link netlink.Link) (err error) {
	tasks := []addressTask{}

	ifaceAddrs := make([]IPNet, 0)
	for _, peer := range iface.Peers {
		for _, ip := range peer.AllowedIPs {
			ifaceAddrs = append(ifaceAddrs, ip)
		}
	}

	deviceAddrs_, err := a.handle.RouteList(link, netlink.FAMILY_ALL)
	if err != nil {
		return fmt.Errorf("listing routes on wg interface: %w", err)
	}
	deviceAddrs := make([]IPNet, len(deviceAddrs_))
	for i, addr := range deviceAddrs_ {
		deviceAddrs[i] = IPNet(*addr.Dst)
	}

	removedIPs := setDifference(deviceAddrs, ifaceAddrs, lessIPNet)
	for _, ip := range removedIPs {
		tasks = append(tasks, addressTask{add: false, ip: ip})
	}
	addedIPs := setDifference(ifaceAddrs, deviceAddrs, lessIPNet)
	for _, ip := range addedIPs {
		tasks = append(tasks, addressTask{add: true, ip: ip})
	}

	tasksStrings := make([]string, len(tasks))
	for i, task := range tasks {
		tasksStrings[i] = task.String()
	}
	zap.S().Debugf("changing %d routes to wg interface:\n%s", len(tasks), strings.Join(tasksStrings, "\n"))

	for _, task := range tasks {
		if task.add {
			err = a.handle.RouteAdd(&netlink.Route{
				LinkIndex: link.Attrs().Index,
				Dst:       (*net.IPNet)(&task.ip),
			})
		} else {
			err = a.handle.RouteDel(&netlink.Route{
				LinkIndex: link.Attrs().Index,
				Dst:       (*net.IPNet)(&task.ip),
			})
		}
		if err != nil {
			return fmt.Errorf("task to wg interface (%s) failed: %w", task, err)
		}
	}
	return nil
}

type addressTask struct {
	add bool
	ip  IPNet
}

func (r addressTask) String() string {
	if r.add {
		return fmt.Sprintf("+ %s", r.ip)
	}
	return fmt.Sprintf("- %s", r.ip)
}

func (a *Applier) configureWireguard(iface Interface) error {
	var err error
	peers := make([]wgtypes.PeerConfig, len(iface.Peers))
	for i, peer := range iface.Peers {
		var endpoint *net.UDPAddr
		if peer.Endpoint != "" {
			zap.S().Debugf("resolving %s for peer %s.", peer.Endpoint, peer.Name)
			endpoint, err = net.ResolveUDPAddr("udp", peer.Endpoint)
			if err != nil {
				return fmt.Errorf("resolving %s for peer %s: %w", peer.Endpoint, peer.Name, err)
			}
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
		PrivateKey:   (*wgtypes.Key)(&iface.PrivateKey),
		ReplacePeers: true,
		Peers:        peers,
	}
	if iface.ListenPort == 0 {
		cfg.ListenPort = nil
	} else {
		cfg.ListenPort = &iface.ListenPort
	}
	zap.S().Debugf("wg interface configuration:\n%s", StringConfig(&cfg))
	err = a.client.ConfigureDevice(iface.Name, cfg)
	if err != nil {
		return fmt.Errorf("configuring wg interface: %w", err)
	}
	zap.S().Debug("wg interface configured.")
	return nil
}
