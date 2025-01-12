package spec

import (
	"fmt"
	"slices"

	"github.com/nyiyui/qrystal/goal"
	"go.uber.org/zap"
)

func (sc SpecCensored) CompileMachine(name string, ignoreIncomplete bool) (goal.Machine, error) {
	gm := goal.Machine{}
	for _, sn := range sc.Networks {
		sndI := slices.IndexFunc(sn.Devices, func(snd NetworkDeviceCensored) bool { return snd.Name == name })
		if sndI == -1 {
			continue
		}
		snd := sn.Devices[sndI]
		for _, name := range snd.Accessible {
			forwardee, ok := sn.GetDevice(name)
			if !ok {
				panic("malformed spec")
			}
			for _, addr := range forwardee.Addresses {
				if addr.IP.To4() != nil {
					// assume IPv6 addresses that represent IPv4 addresses just need the v4 option :)
					// ^ idk if this is true
					gm.ForwardsIPv4 = true
				} else {
					// assume (not IPv4) → IPv6
					gm.ForwardsIPv6 = true
				}
			}
		}
		forwardsFor := make([][]int, len(sn.Devices))
		for i, snd := range sn.Devices {
			if i == sndI {
				continue
			}
			if snd.PublicKey == (goal.Key{}) {
				continue
			}
			if snd.ForwarderAndEndpointChosen && snd.UsesForwarder {
				forwardsFor[snd.ForwarderChosenIndex] = append(forwardsFor[snd.ForwarderChosenIndex], i)
			}
		}
		peers := make([]goal.InterfacePeer, 0, len(sn.Devices)-1)
		for i, snd := range sn.Devices {
			if i == sndI {
				continue
			}
			if snd.PublicKey == (goal.Key{}) {
				if ignoreIncomplete {
					zap.S().Debugf("%s/%s has unset PublicKey, ignore.", sn.Name, snd.Name)
					continue
				} else {
					return goal.Machine{}, fmt.Errorf("%s/%s has unset PublicKey", sn.Name, snd.Name)
				}
			}
			var endpoint string
			if !snd.ForwarderAndEndpointChosen {
				zap.S().Debugf("%s/%s does not have a chosen forwarder and endpoint, proceed with blank Endpoint.", sn.Name, snd.Name)
			} else {
				if !snd.UsesForwarder {
					endpoint = snd.Endpoints[snd.EndpointChosenIndex]
				} else {
					forwarder := sn.Devices[snd.ForwarderChosenIndex]
					if !forwarder.ForwarderAndEndpointChosen {
						// NOTE: this shouldn't happen; the forwarder should have been checked before this
						if ignoreIncomplete {
							zap.S().Debugf("%s/%s has forwarder %s/%s which does not have a chosen forwarder and endpoint, ignore.", sn.Name, snd.Name, sn.Name, forwarder.Name)
							continue
						} else {
							return goal.Machine{}, fmt.Errorf("%s/%s has forwarder %s/%s which does not have a chosen forwarder and endpoint", sn.Name, snd.Name, sn.Name, forwarder.Name)
						}
					}
					continue
				}
			}
			allowedIPs := snd.Addresses
			thisForwardsFor := forwardsFor[i]
			for _, j := range thisForwardsFor {
				forwardee := sn.Devices[j]
				if forwardee.PublicKey == (goal.Key{}) {
					if ignoreIncomplete {
						zap.S().Debugf("%s/%s has forwarder %s/%s which has unset PublicKey, ignore.", sn.Name, snd.Name, sn.Name, forwardee.Name)
						continue
					} else {
						return goal.Machine{}, fmt.Errorf("%s/%s has forwarder %s/%s which has unset PublicKey", sn.Name, snd.Name, sn.Name, forwardee.Name)
					}
				}
				allowedIPs = append(allowedIPs, forwardee.Addresses...)
			}
			peers = append(peers, goal.InterfacePeer{
				Name:                snd.Name,
				PublicKey:           snd.PublicKey,
				PresharedKey:        snd.PresharedKey,
				Endpoint:            endpoint,
				PersistentKeepalive: snd.PersistentKeepalive,
				AllowedIPs:          allowedIPs,
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
