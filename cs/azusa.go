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
	var desc strings.Builder
	for cnn, peer := range q.Networks {
		err = checkPeer(ti, cnn, peer)
		if err != nil {
			return err
		}
		_, ok := c.cc.Networks[cnn]
		if !ok {
			return fmt.Errorf("net %s no exist :(", cnn)
		}
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
	c.ccLock.Lock()
	defer c.ccLock.Unlock()
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
		cn.Peers[peer.Name] = &central.Peer{
			Hosts:      peer.Hosts,
			AllowedIPs: peer.AllowedIPs,
			CanSee:     peer.CanSee,
		}
	}
	return nil
}
