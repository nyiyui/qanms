package goal

import (
	"bytes"
	"net"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func mustParseCIDR(addr string) IPNet {
	_, ret, err := net.ParseCIDR(addr)
	if err != nil {
		panic(err)
	}
	return IPNet(*ret)
}

func TestDiffInterfacePeer(t *testing.T) {
	a := InterfacePeer{
		PublicKey:           Key{1},
		PresharedKey:        nil,
		Endpoint:            "localhost",
		PersistentKeepalive: 0,
		AllowedIPs: []IPNet{
			mustParseCIDR("10.10.0.0/24"),
			mustParseCIDR("10.10.1.0/25"),
		},
	}
	b := InterfacePeer{
		PublicKey:           Key{0},
		PresharedKey:        nil,
		Endpoint:            "127.0.0.1",
		PersistentKeepalive: Duration(1 * time.Second),
		AllowedIPs: []IPNet{
			mustParseCIDR("10.10.1.0/25"),
			mustParseCIDR("10.10.2.0/32"),
		},
	}
	got := DiffInterfacePeer(&a, &b)
	want := InterfacePeerDiff{
		PublicKeyChanged:           true,
		EndpointChanged:            true,
		PersistentKeepaliveChanged: true,
		AllowedIPsChanged: []Change[IPNet]{
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
