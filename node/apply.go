package node

import (
	"github.com/nyiyui/qrystal/central"
	"github.com/nyiyui/qrystal/util"
	"golang.org/x/exp/slices"
)

// applyCC applies cc2 to n.
// ccLock is locked by the caller
func (n *Node) applyCC(cc2 *central.Config) {
	// NOTE: shouldn't have to lock any more since ccLock overrides all inner locks (i.e. inner locks have to be used in conjunction with ccLock)
	if n.cc.Networks == nil {
		n.cc.Networks = map[string]*central.Network{}
		n.cc.Desynced |= central.DNetwork
	}
	for cnn2, cn2 := range cc2.Networks {
		cn, ok := n.cc.Networks[cnn2]
		if !ok {
			// new cn
			util.S.Debugf("apply2 new CN %s %#v", cnn2, cn)
			n.cc.Networks[cnn2] = &central.Network{}
			cn = n.cc.Networks[cnn2]
			n.cc.Desynced |= central.DNetwork
		}
		util.S.Debugf("apply2 cn %s %#v", cnn2, cn)
		cn.Desynced = 0
		cn.Name = cnn2
		cn.IPs = cn2.IPs
		forwardingPeers := map[string]struct{}{}
		if cn.Peers == nil {
			cn.Peers = map[string]*central.Peer{}
			cn.Desynced |= central.DIPs
		}
		if !central.Same(cn.IPs, cn2.IPs) || !central.Same2(cn.Peers, cn2.Peers) {
			cn.Desynced |= central.DIPs
		}
		for pn2, peer2 := range cn2.Peers {
			peer, ok := cn.Peers[pn2]
			if !ok {
				// new peer
				peer2.Name = pn2
				cn.Peers[pn2] = peer2
				cn.Desynced |= central.DIPs
				continue
			}
			peer.Desynced = 0
			if !peer.Same(peer2) {
				peer.Desynced = central.DPeer
			}
			peer.Name = pn2
			peer.Host = peer2.Host
			peer.CanForward = peer2.CanForward
			peer.AllowedIPs = peer2.AllowedIPs
			if !central.Same(peer.AllowedIPs, peer2.AllowedIPs) {
				peer.Desynced |= central.DIPs
			}
			peer.ForwardingPeers = []string{}
			for _, forwardingPeer := range peer2.ForwardingPeers {
				if forwardingPeer == cn.Me {
					continue
				}
				_, ok := forwardingPeers[forwardingPeer]
				if !ok {
					peer.ForwardingPeers = append(peer.ForwardingPeers, forwardingPeer)
					forwardingPeers[forwardingPeer] = struct{}{}
				}
			}
			peer.CanSee = peer2.CanSee
			if peer.PubKey != peer2.PubKey {
				peer.Desynced |= central.DKeys
			}
			peer.PubKey = peer2.PubKey
			if peer.PubKey != peer2.PubKey {
				peer.Desynced |= central.DKeys
			}
			peer.SRVs = peer2.SRVs
			if !slices.Equal(peer.SRVs, peer2.SRVs) {
				peer.Desynced |= central.DSRVs
			}
			cn.Peers[pn2] = peer
		}
		for pn := range cn.Peers {
			cn.Desynced |= central.DIPs
			_, ok := cn2.Peers[pn]
			if !ok {
				// removed
				delete(cn.Peers, pn)
			}
		}
		cn.Me = cn2.Me
		cn.Keepalive = cn2.Keepalive
		cn.ListenPort = cn2.ListenPort
		n.cc.Networks[cnn2] = cn
	}
	for cnn := range n.cc.Networks {
		_, ok := cc2.Networks[cnn]
		if !ok {
			// removed
			delete(n.cc.Networks, cnn)
		}
		n.cc.Desynced |= central.DNetwork
	}
}
