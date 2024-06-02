package goal

import (
	"bytes"
	"net"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func mustParseCIDR(addr string) net.IPNet {
	_, ret, err := net.ParseCIDR(addr)
	if err != nil {
		panic(err)
	}
	return *ret
}

func TestDiffInterfacePeer(t *testing.T) {
	a := InterfacePeer{
		PublicKey:           wgtypes.Key{1},
		PresharedKey:        wgtypes.Key{1},
		Endpoint:            "localhost",
		PersistentKeepalive: 0,
		AllowedIPs: []net.IPNet{
			mustParseCIDR("10.10.0.0/24"),
			mustParseCIDR("10.10.1.0/25"),
		},
	}
	b := InterfacePeer{
		PublicKey:           wgtypes.Key{0},
		PresharedKey:        wgtypes.Key{1},
		Endpoint:            "127.0.0.1",
		PersistentKeepalive: 1 * time.Second,
		AllowedIPs: []net.IPNet{
			mustParseCIDR("10.10.1.0/25"),
			mustParseCIDR("10.10.2.0/32"),
		},
	}
	got := DiffInterfacePeer(&a, &b)
	want := InterfacePeerDiff{
		PublicKeyChanged:           true,
		EndpointChanged:            true,
		PersistentKeepaliveChanged: true,
		AllowedIPsChanged: []Change[net.IPNet]{
			{ChangeOpRemove, mustParseCIDR("10.10.0.0/24")},
			{ChangeOpNoChange, mustParseCIDR("10.10.1.0/25")},
			{ChangeOpAdd, mustParseCIDR("10.10.2.0/32")},
		},
		AllowedIPsNoChange: false,
	}
	sort.Slice(got.AllowedIPsChanged, func(i, j int) bool {
		return bytes.Compare(got.AllowedIPsChanged[i].Value.IP, got.AllowedIPsChanged[j].Value.IP) < 0
	})
	sort.Slice(want.AllowedIPsChanged, func(i, j int) bool {
		return bytes.Compare(want.AllowedIPsChanged[i].Value.IP, want.AllowedIPsChanged[j].Value.IP) < 0
	})
	if !cmp.Equal(got, want) {
		t.Log(cmp.Diff(got, want))
		t.Fatal("mismatch")
	}
}
