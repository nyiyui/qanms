package goal

import (
	"bytes"
	"net"
	"testing"
)

func TestSetDifference(t *testing.T) {
	type test struct {
		a, b, want []string
	}
	tests := []test{
		{[]string{"10.0.0.1/32", "10.0.0.2/32"}, []string{"10.0.0.1/32"}, []string{"10.0.0.2/32"}},
		{[]string{"10.0.0.1/32", "10.0.0.2/32"}, []string{"10.0.0.2/32"}, []string{"10.0.0.1/32"}},
		{[]string{"10.0.0.1/32", "10.0.0.2/32"}, []string{"10.0.0.3/32"}, []string{"10.0.0.1/32", "10.0.0.2/32"}},
		{[]string{"10.0.0.1/32", "10.0.0.2/32"}, []string{"10.0.0.1/32", "10.0.0.2/32"}, []string{}},
		{[]string{"10.0.0.1/32"}, []string{"10.0.0.1/32", "10.0.0.2/32"}, []string{}},
	}
	mustParseCIDR := func(s string) IPNet {
		_, ipnet, err := net.ParseCIDR(s)
		if err != nil {
			t.Fatalf("failed to parse CIDR %q: %v", s, err)
		}
		return IPNet(*ipnet)
	}
	for _, tt := range tests {
		a := make([]IPNet, len(tt.a))
		for i, s := range tt.a {
			a[i] = mustParseCIDR(s)
		}
		b := make([]IPNet, len(tt.b))
		for i, s := range tt.b {
			b[i] = mustParseCIDR(s)
		}
		want := make([]IPNet, len(tt.want))
		for i, s := range tt.want {
			want[i] = mustParseCIDR(s)
		}
		got := setDifference(a, b)
		if len(got) != len(want) {
			t.Errorf("setDifference(%v, %v) = %v; want %v", a, b, got, want)
			continue
		}
		for i, want := range want {
			if !bytes.Equal(got[i].IP, want.IP) || !bytes.Equal(got[i].Mask, want.Mask) {
				t.Errorf("setDifference(%v, %v) = %v; want %v", a, b, got, want)
			}
		}
	}
}
