package spec

import (
	"fmt"
	"slices"

	"github.com/nyiyui/qrystal/goal"
)

func (sc SpecCensored) CompileMachine(name string, ignoreIncomplete bool) (goal.Machine, error) {
	gm := goal.Machine{}
	for _, sn := range sc.Networks {
		sndI := slices.IndexFunc(sn.Devices, func(snd NetworkDeviceCensored) bool { return snd.Name == name })
		if sndI == -1 {
			continue
		}
		snd := sn.Devices[sndI]
		peers := make([]goal.InterfacePeer, 0, len(sn.Devices)-1)
		for i, snd := range sn.Devices {
			if i == sndI {
				continue
			}
			if !snd.EndpointChosen {
				if ignoreIncomplete {
					continue
				} else {
					return goal.Machine{}, fmt.Errorf("%s/%s does not have a chosen endpoint", sn.Name, snd.Name)
				}
			}
			if snd.PublicKey == (goal.Key{}) {
				if ignoreIncomplete {
					continue
				} else {
					return goal.Machine{}, fmt.Errorf("%s/%s has unset PublicKey", sn.Name, snd.Name)
				}
			}
			peers = append(peers, goal.InterfacePeer{
				Name:                snd.Name,
				PublicKey:           snd.PublicKey,
				PresharedKey:        snd.PresharedKey,
				Endpoint:            snd.Endpoints[snd.EndpointChosenIndex],
				PersistentKeepalive: snd.PersistentKeepalive,
				AllowedIPs:          snd.Addresses,
				// TODO: forwarding: add forwarding addresses to goal.InterfacePeer.AllowedIPs
			})
		}
		gm.Interfaces = append(gm.Interfaces, goal.Interface{
			Name:       sn.Name,
			ListenPort: snd.ListenPort,
			Addresses:  snd.Addresses,
			Peers:      peers,
		})
	}
	return gm, nil
}
