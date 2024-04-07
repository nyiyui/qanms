package cs

import (
	"fmt"
	"strings"

	"github.com/cenkalti/rpc2"
	"github.com/nyiyui/qrystal/api"
	"github.com/nyiyui/qrystal/central"
	"github.com/nyiyui/qrystal/util"
)

func (c *CentralSource) azusa(cl *rpc2.Client, q *api.AzusaQ, s *api.AzusaS) error {
	ti, ok, err := c.Tokens.getToken(&q.CentralToken)
	if err != nil {
		return err
	}
	if !ok {
		return newTokenAuthError(q.CentralToken)
	}
	c.ccLock.Lock()
	defer c.ccLock.Unlock()
	var desc strings.Builder
	for cnn, peer := range q.Networks {
		err = checkPeer(ti, cnn, c.cc, peer)
		if err != nil {
			return err
		}
		_, ok = c.cc.Networks[cnn]
		if !ok {
			return fmt.Errorf("net %s no exist :(", cnn)
		}
		if !ti.SRVAllowancesAny {
			for saI, sa := range q.Networks[cnn].AllowedSRVs {
				if !central.AllowedByAny(sa, ti.SRVAllowances) {
					return fmt.Errorf("peer allowance %d: not allowed by any token-level allowances", saI)
				}
			}
		}
		// TODO: token-level restrictions on SRVAllowance and SRVs
		fmt.Fprintf(&desc, "\n- net %s peer %s: %#v", cnn, peer.Name, peer)
	}
	util.S.Infof("azusa from token %s to push %d:\n%s", ti.Name, len(q.Networks), &desc)
	ti.StartUse()
	err = c.Tokens.UpdateToken(ti)
	if err != nil {
		return err
	}
	defer func() {
		ti.StopUse()
		err = c.Tokens.UpdateToken(ti)
		if err != nil {
			util.S.Errorf("UpdateToken %s: %s", ti.key, err)
		}
	}()
	changed := map[string][]string{}
	for cnn, peer := range q.Networks {
		cn := c.cc.Networks[cnn]
		if peer.AllowedIPs == nil || len(peer.AllowedIPs) == 0 {
			ipNet, err := cn.AssignAddr()
			if err != nil {
				return err
			}
			util.S.Infof("azusa from token %s to push net %s peer %s: assign IP %#v", ti.Name, cnn, peer.Name, ipNet)
			peer.AllowedIPs = []central.IPNet{
				central.IPNet(ipNet),
			}
		}
		prevPeer := cn.Peers[peer.Name]
		cn.Peers[peer.Name] = &central.Peer{
			Host:             peer.Host,
			AlternativeHosts: peer.AlternativeHosts,
			AllowedIPs:       peer.AllowedIPs,
			CanForward:       peer.CanForward,
			CanSee:           peer.CanSee,
			AllowedSRVs:      peer.AllowedSRVs,
		}
		if len(prevPeer.SRVs) > 0 {
			okSRVs := make([]central.SRV, 0)
			for _, srv := range prevPeer.SRVs {
				if !central.AllowedByAny(srv, peer.AllowedSRVs) {
					continue
				}
				okSRVs = append(okSRVs, srv)
			}
			if len(okSRVs) < len(prevPeer.SRVs) {
				util.S.Infof("azusa from token %s to push net %s peer %s: previous SRV records not allowed by new AllowedSRVs rule; removing disallowed SRV records (%d previous, %d now disallowed, %d still allowed)", ti.Name, cnn, peer.Name, len(prevPeer.SRVs), len(prevPeer.SRVs)-len(okSRVs), len(okSRVs))
			}
			cn.Peers[peer.Name].SRVs = okSRVs
		}
		changed[cnn] = []string{peer.Name}
	}
	c.notify(change{Reason: fmt.Sprintf("azusa %s", ti.Name), Changed: changed})
	return nil
}
