package node

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"strings"

	"github.com/cenkalti/rpc2"
	"github.com/nyiyui/qrystal/api"
	"github.com/nyiyui/qrystal/cs"
	"github.com/nyiyui/qrystal/util"
	"golang.org/x/exp/slices"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func (n *Node) setupClient(cl *rpc2.Client) {
	cl.Handle("ping", func(cl *rpc2.Client, q *bool, s *bool) error {
		*s = true
		return nil
	})
	cl.Handle("keepalive", func(cl *rpc2.Client, q *api.KeepaliveQ, s *api.KeepaliveS) error {
		n.keepaliveEntriesLock.Lock()
		defer n.keepaliveEntriesLock.Unlock()
		i := slices.IndexFunc(n.keepaliveEntries, func(ke keepaliveEntry) bool { return bytes.Equal(ke.Secret, *q) })
		if i == -1 {
			return errors.New("keepalive entry not found")
		}
		n.keepaliveEntries[i].Notify <- struct{}{}
		return nil
	})
	cl.Handle("push", func(cl *rpc2.Client, q *api.PushQ, s *api.PushS) error {
		// TODO: notify ERRNO= for errors?
		var err error
		cc := q.CC

		util.Notify(fmt.Sprintf("STATUS=Receiving new CC (%d CNs)...", len(cc.Networks)))
		for cnn, cn := range cc.Networks {
			util.S.Debugf("新たなCCを受信: net %s: %s", cnn, cn)
		}

		// filter out NetworksAllowed
		for cnn, cn := range cc.Networks {
			if !n.cs.netAllowed(cnn) {
				delete(cc.Netowrks, cnn)
				util.S.Warnf("net not allowed; ignoring: %s", cnn)
			}
		}

		n.ccLock.Lock()
		defer n.ccLock.Unlock()
		err = n.removeDevices(cs.MissingFromFirst(cc.Networks, n.cc.Networks))
		if err != nil {
			return fmt.Errorf("rm devs: %w", err)
		}
		n.applyCC(&cc)
		for cnn, cn := range n.cc.Networks {
			util.S.Debugf("after applyCC: net %s: %s", cnn, cn)
		}
		s.PubKeys = map[string]wgtypes.Key{}
		for cnn, cn := range n.cc.Networks {
			if cn.MyPrivKey == nil {
				key, err := wgtypes.GeneratePrivateKey()
				if err != nil {
					return fmt.Errorf("GeneratePrivateKey: %w", err)
				}
				cn.MyPrivKey = &key
				n.cc.Networks[cnn] = cn
				util.S.Infof("net %s: my *new* PublicKey is %s", cnn, key.PublicKey())
			}
			s.PubKeys[cnn] = cn.MyPrivKey.PublicKey()
		}
		return n.update(updateModeSynced)
	})
	go cl.Run()
}

func (n *Node) newClient(ctx context.Context) (*rpc2.Client, *tls.Conn, error) {
	var tlsCfg *tls.Config
	if n.cs.TLSConfig != nil {
		tlsCfg = n.cs.TLSConfig.Clone()
	}
	conn, err := (&tls.Dialer{Config: tlsCfg}).DialContext(ctx, "tcp", n.cs.Host)
	if err != nil {
		err = fmt.Errorf("dial: %w", err)
		return nil, nil, err
	}
	cl := rpc2.NewClient(conn)
	n.setupClient(cl)
	var b bool
	err = cl.CallWithContext(ctx, "ping", true, &b)
	if err != nil {
		err = fmt.Errorf("ping: %w", err)
		return cl, conn.(*tls.Conn), fmt.Errorf("ping: %s", err)
	}
	return cl, conn.(*tls.Conn), nil
}

type updateMode int

const (
	updateModeSynced updateMode = iota + 1
	updateModeBackport
)

func (u updateMode) String() string {
	switch u {
	case updateModeSynced:
		return "synced"
	case updateModeBackport:
		return "backported"
	default:
		return fmt.Sprintf("updateMode unknown %d", u)
	}
}

func (n *Node) update(mode updateMode) error {
	util.Notify(fmt.Sprintf("STATUS=Reifying %s CC (%d CNs)...", mode, len(n.cc.Networks)))
	err := n.reify()
	if err != nil {
		return fmt.Errorf("reify: %w", err)
	}
	util.Notify(fmt.Sprintf("STATUS=Updating DNS for %s CC (%d CNs)...", mode, len(n.cc.Networks)))
	err = n.updateHokutoCC()
	if err != nil {
		return fmt.Errorf("updateHokutoCC: %w", err)
	}
	cns := new(strings.Builder)
	for cnn := range n.cc.Networks {
		fmt.Fprintf(cns, " %s", cnn)
	}
	util.Notify(fmt.Sprintf("STATUS=OK: %s CNs:%s", mode, cns))
	util.Notify("READY=1")
	err = n.saveBackport()
	if err != nil {
		util.S.Errorf("save backport: %s", err)
	}
	n.traceCheck()
	return nil
}
