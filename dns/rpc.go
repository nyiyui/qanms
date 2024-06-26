package dns

import (
	"net/rpc"
	"sync"

	"github.com/nyiyui/qrystal/spec"
	"go.uber.org/zap"
)

type RPCServer struct {
	spec     *spec.SpecCensored
	specLock sync.RWMutex
}

func (r *RPCServer) UpdateSpec(spec spec.SpecCensored, alwaysNil *bool) error {
	r.specLock.Lock()
	defer r.specLock.Unlock()
	r.spec = &spec
	zap.S().Info("updated spec.")
	return nil
}

type Client interface {
	// UpdateSpec sends an updated spec to the DNS server.
	UpdateSpec(spec.SpecCensored) error
}

type DirectClient struct {
	r *RPCServer
}

func NewDirectClient(s *Server) *DirectClient {
	return &DirectClient{r: s.r}
}

// UpdateSpec sends an updated spec to the DNS server.
func (d *DirectClient) UpdateSpec(spec spec.SpecCensored) error {
	return d.r.UpdateSpec(spec, new(bool))
}

// RPCClient is the client for RPCServer.
type RPCClient struct {
	c *rpc.Client
}

func NewRPCClient(c *rpc.Client) *RPCClient {
	return &RPCClient{c: c}
}

// UpdateSpec sends an updated spec to the DNS server.
func (r *RPCClient) UpdateSpec(spec spec.SpecCensored) error {
	return r.c.Call("RPCServer.UpdateSpec", spec, new(bool))
}

// Close calls the underlying rpc.Client.Close.
func (r *RPCClient) Close() error {
	return r.c.Close()
}
