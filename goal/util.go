package goal

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type IPNet net.IPNet

func (in *IPNet) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	_, in2, err := net.ParseCIDR(s)
	if err != nil {
		return fmt.Errorf("parsing CIDR: %w", err)
	}
	in.IP = in2.IP
	in.Mask = in2.Mask
	return nil
}

func (in *IPNet) MarshalJSON() ([]byte, error) {
	return json.Marshal((*net.IPNet)(in).String())
}

func ipNetUtilToStd(s []IPNet) []net.IPNet {
	s2 := make([]net.IPNet, len(s))
	for i := range s {
		s2[i] = net.IPNet(s[i])
	}
	return s2
}

type Key wgtypes.Key

func (k *Key) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	if s == "" {
		return nil
	}
	k2, err := wgtypes.ParseKey(s)
	if err != nil {
		return err
	}
	*k = Key(k2)
	return nil
}

func (k Key) MarshalJSON() ([]byte, error) {
	return json.Marshal(wgtypes.Key(k).String())
}

func StringConfig(cfg *wgtypes.Config) string {
	b := new(strings.Builder)
	tags := ""
	if cfg.ReplacePeers {
		tags += " replace"
	}
	fmt.Fprintf(b, "[Interface]%s\n", tags)
	fmt.Fprintf(b, "PrivateKey = %s\n", cfg.PrivateKey)
	fmt.Fprintf(b, "ListenPort = %d\n", cfg.ListenPort)
	fmt.Fprintf(b, "FirewallMark = %d\n", cfg.FirewallMark)
	fmt.Fprintf(b, "ReplacePeers = %t\n\n", cfg.ReplacePeers)
	for i, peer := range cfg.Peers {
		tags := ""
		if peer.Remove {
			tags += " remove"
		}
		if peer.UpdateOnly {
			tags += " update-only"
		}
		if peer.ReplaceAllowedIPs {
			tags += " replace-allowed-ips"
		}
		fmt.Fprintf(b, "[Peer %d]%s\n", i, tags)
		fmt.Fprintf(b, "PublicKey = %s\n", peer.PublicKey)
		fmt.Fprintf(b, "PresharedKey = %s\n", peer.PresharedKey)
		fmt.Fprintf(b, "Keepalive = %d\n", peer.PersistentKeepaliveInterval)
		fmt.Fprintf(b, "Endpoint = %s\n", peer.Endpoint)
		fmt.Fprintf(b, "AllowedIPs = %v\n", peer.AllowedIPs)
	}
	return b.String()
}
