package dns

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"slices"
	"strings"

	"github.com/miekg/dns"
	"github.com/nyiyui/qrystal/goal"
	"github.com/nyiyui/qrystal/util"
	"go.uber.org/zap"
)

var mask32 = net.CIDRMask(32, 32)
var mask128 = net.CIDRMask(128, 128)

// ~~stolen~~ copied from <https://gist.github.com/walm/0d67b4fb2d5daf3edd4fad3e13b162cb>.

type Server struct {
	r       *RPCServer
	parents []Parent
}

func NewServer(parents []Parent) (*Server, error) {
	if len(parents) == 0 {
		return nil, errors.New("must have at least one parent")
	}
	// === check config ===
	suffixes := map[string]int{}
	for i, parent := range parents {
		if parent.Suffix == "" {
			return nil, fmt.Errorf("parent index %d: no suffix", i)
		}
		if parent.Suffix[len(parent.Suffix)-1] == '.' {
			return nil, fmt.Errorf("parent index %d: suffix must not end with a period (trailing period should not be used)", i)
		}
		if (parent.Network == "") != (parent.Device == "") {
			return nil, fmt.Errorf("parent index %d: must have none or both of Network and Device", i)
		}
		if parent.Network != "" && parent.Device != "" && parent.Suffix[0] == '.' {
			return nil, fmt.Errorf("parent index %d: Suffix must NOT start with a dot if Network and Device is specified", i)
		}
		if parent.Network == "" && parent.Device == "" && parent.Suffix[0] != '.' {
			return nil, fmt.Errorf("parent index %d: Suffix must start with a dot if Network and Device is specified", i)
		}
		if j, ok := suffixes[parent.Suffix]; ok {
			return nil, fmt.Errorf("parent index %d and %d have duplicate suffixes", i, j)
		}
		suffixes[parent.Suffix] = i
	}

	return &Server{
		r:       new(RPCServer),
		parents: parents,
	}, nil
}

type Parent struct {
	Suffix  string
	Network string
	Device  string
}

func (s *Server) ListenDNS(addr string) error {
	mux := dns.NewServeMux()
	mux.HandleFunc(".", s.handle)
	server := &dns.Server{Addr: addr, Net: "udp", Handler: mux}
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			util.S.Fatalf("dns ListenAndServe failed: %s\n ", err.Error())
		}
	}()
	return nil
}

func (s *Server) ListenRPC(socketPath string) error {
	lis, err := net.Listen("unix", socketPath)
	if err != nil {
		return err
	}
	rs := rpc.NewServer()
	rs.Register(s.r)
	go rs.Accept(lis)
	return nil
}

func (s *Server) handle(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false
	switch r.Opcode {
	case dns.OpcodeQuery:
		m.MsgHdr.Rcode = s.handleQuery(m)
	}
	w.WriteMsg(m)
}

func (s *Server) handleQuery(r *dns.Msg) (rcode int) {
	for _, q := range r.Question {
		for _, parent := range s.parents {
			if strings.HasSuffix(q.Name[:len(q.Name)-1], parent.Suffix) {
				switch q.Qtype {
				case dns.TypeA:
					rcode2 := s.handleParent(r, q, parent)
					if rcode2 != dns.RcodeSuccess {
						rcode = rcode2
						return
					}
				}
				break
			}
		}
	}
	return dns.RcodeSuccess
}

func (s *Server) handleParent(r *dns.Msg, q dns.Question, parent Parent) int {
	s.r.specLock.RLock()
	defer s.r.specLock.RUnlock()
	if s.r.spec == nil {
		util.S.Error("spec is not available.")
		return dns.RcodeServerFailure
	}
	parts := strings.Split(strings.TrimSuffix(q.Name[:len(q.Name)-1], parent.Suffix), ".")
	if len(parts) == 1 && parts[0] == "" {
		parts = nil
	}
	slices.Reverse(parts)
	if parent.Device != "" {
		parts = append([]string{parent.Device}, parts...)
	}
	if parent.Network != "" {
		parts = append([]string{parent.Network}, parts...)
	}
	zap.S().Debugf("parts is %#v.", parts)
	if len(parts) != 2 {
		return dns.RcodeNameError
	}
	network, device := parts[0], parts[1]
	nc, ok := s.r.spec.GetNetwork(network)
	if !ok {
		zap.S().Debugf("%s/ not found", network)
		return dns.RcodeNameError
	}
	ndc, ok := nc.GetDevice(device)
	if !ok {
		zap.S().Debugf("%s/%s not found", network, device)
		return dns.RcodeNameError
	}
	s.returnAddresses(r, q, ndc.Addresses)
	return dns.RcodeSuccess
}

func (s *Server) returnAddresses(r *dns.Msg, q dns.Question, addresses []goal.IPNet) {
	for _, ipNet := range addresses {
		if bytes.Equal(ipNet.Mask, mask32) {
			rr := dns.A{
				Hdr: dns.RR_Header{
					Name:   q.Name,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    0,
				},
				A: ipNet.IP,
			}
			r.Answer = append(r.Answer, &rr)
		} else if bytes.Equal(ipNet.Mask, mask128) {
			rr := dns.AAAA{
				Hdr: dns.RR_Header{
					Name:   q.Name,
					Rrtype: dns.TypeAAAA,
					Class:  dns.ClassINET,
					Ttl:    0,
				},
				AAAA: ipNet.IP,
			}
			r.Answer = append(r.Answer, &rr)
		} else {
			zap.S().Errorf("%s/%s has non-/32 or non-/128 IP address, ignoring in DNS response.")
			continue
		}
	}
}
