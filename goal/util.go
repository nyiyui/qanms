package goal

import (
	"encoding/json"
	"fmt"
	"net"

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
