package node

import (
	"fmt"

	"github.com/nyiyui/qrystal/hokuto"
	"github.com/nyiyui/qrystal/util"
)

// updateHokutoCC updates hokuto's copy of CC.
// NOTE: Node.ccLock must be locked!
func (n *Node) updateHokutoCC() error {
	for cnn, cn := range n.cc.Networks {
		util.S.Debugf("updateHokutoCC pre: net %s: %s", cnn, cn)
	}
	var dummy bool
	q := hokuto.UpdateCCQ{
		Token: n.hokuto.token,
		CC:    &n.cc,
	}
	err := n.hokuto.client.Call("Hokuto.UpdateCC", q, &dummy)
	if err != nil {
		return fmt.Errorf("call: %w", err)
	}
	for cnn, cn := range n.cc.Networks {
		util.S.Debugf("updateHokutoCC post: net %s: %s", cnn, cn)
	}
	return nil
}

func (n *Node) hokutoInit(parent, addr string, extraParents []hokuto.ExtraParent) error {
	var dummy bool
	q := hokuto.InitQ{
		Parent:       parent,
		Addr:         addr,
		ExtraParents: extraParents,
	}
	err := n.hokuto.client.Call("Hokuto.Init", q, &dummy)
	if err != nil {
		return fmt.Errorf("call: %w", err)
	}
	return nil
}
