package cs

import (
	"errors"
	"fmt"
	"net"

	"github.com/nyiyui/qrystal/central"
)

type httpError struct {
	code int
	err  error
}

func newHttpError(code int, err error) httpError {
	return httpError{code, err}
}
func newHttpErrorf(code int, format string, a ...any) httpError {
	return httpError{code, fmt.Errorf(format, a...)}
}

func (h httpError) Error() string {
	return fmt.Sprintf("%s (http error code %d)", h.err.Error(), h.code)
}

func (h httpError) Unwrap() error { return h.err }

func checkPeer(ti TokenInfo, cnn string, cc central.Config, peer central.Peer) error {
	cn := cc.Networks[cnn]
	if ti.CanPush == nil {
		return newHttpError(403, errors.New("token: cannot push"))
	}
	if !ti.CanPush.Any {
		prelude := fmt.Sprintf("cannot push to net %s peer %s", cnn, peer.Name)
		cpn, ok := ti.CanPush.Networks[cnn]
		if !ok {
			return newHttpErrorf(403, "%s as token cannot push to this net", prelude)
		}
		if peer.Name != cpn.Name {
			return newHttpErrorf(403, "%s as only peer %s is allowed", prelude, cpn.Name)
		}
		for _, in := range peer.AllowedIPs {
			if !central.IPNetSubsetOfAny(net.IPNet(in), cn.IPs) {
				return newHttpErrorf(403, "%s as peer specifies AllowedIPs not in network's IPs", prelude)
			}
		}
		if cpn.CanSeeElementAny == false {
			if peer.CanSee == nil {
				return newHttpErrorf(403, "%s as peer specifies CanSee any but CanSeeElement does not allow any", prelude)
			} else if len(MissingFromFirst(SliceToMap(cpn.CanSeeElement), SliceToMap(peer.CanSee.Only))) != 0 {
				return newHttpErrorf(403, "%s as peer CanSee specifies (%s) a superset of allowed (%s)", prelude, peer.CanSee.Only, cpn.CanSeeElement)
			}
		}
	}
	return nil
}

func checkSrv(ti TokenInfo, cnn string) error {
	if ti.CanPush == nil {
		return newHttpError(403, errors.New("token: cannot push"))
	}
	if !ti.CanSRVUpdate {
		return newHttpError(403, errors.New("token: cannot srvUpdate"))
	}
	if !ti.CanPush.Any {
		_, ok := ti.CanPush.Networks[cnn]
		if !ok {
			return newHttpErrorf(403, "token cannot push to this net")
		}
	}
	return nil
}
