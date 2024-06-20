package coord

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/nyiyui/qrystal/goal"
	"github.com/nyiyui/qrystal/spec"
	"github.com/nyiyui/qrystal/util"
)

type TokenInfo struct {
	Identities [][2]string
}

type Server struct {
	mux      *http.ServeMux
	spec     spec.Spec
	specLock sync.RWMutex
	latest   map[string][]string
	tokens   map[util.TokenHash]TokenInfo
}

func NewServer(spec spec.Spec, tokens map[util.TokenHash]TokenInfo) *Server {
	if tokens == nil {
		panic("coord.NewServer: tokens map must not be nil")
	}
	s := &Server{
		mux:    http.NewServeMux(),
		spec:   spec,
		tokens: tokens,
	}
	s.setup()
	return s
}

// verifyIdentity verifies if the given request has the credentials to identify as the given network device.
// If this returns true, continue with the request.
// If this returns false, abort the request.
func (s *Server) verifyIdentity(w http.ResponseWriter, r *http.Request, network, device string) (ok bool) {
	const prefix = "QrystalCoordIdentityToken "
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, prefix) {
		http.Error(w, "Authorization header must have type QrystalCoordIdentityToken", 401)
		return false
	}
	token, err := util.ParseToken(strings.TrimPrefix(header, prefix))
	if err != nil {
		http.Error(w, "bad token", 401)
		return false
	}
	tokenHash := token.Hash()
	tokenInfo, ok := s.tokens[*tokenHash]
	if !ok {
		http.Error(w, "not authorized", 401)
		return false
	}
	if !slices.Contains(tokenInfo.Identities, [2]string{network, device}) {
		http.Error(w, "not authorized", 401)
		return false
	}
	return true
}

func (s *Server) setup() {
	s.mux.HandleFunc("GET /v1/reify/{network}/{device}/spec", s.getReifySpec)
	s.mux.HandleFunc("PATCH /v1/reify/{network}/{device}/spec", s.patchReifySpec)
	s.mux.HandleFunc("POST /v1/reify/{network}/{device}/status", s.postReifyStatus)
}

func (s *Server) getReifySpec(w http.ResponseWriter, r *http.Request) {
	network := r.PathValue("network")
	device := r.PathValue("device")
	if !s.verifyIdentity(w, r, network, device) {
		return
	}
	s.specLock.RLock()
	defer s.specLock.RUnlock()
	{
		sn, ok := s.spec.GetNetwork(network)
		if !ok {
			http.Error(w, "network not found", 404)
			return
		}
		if _, ok = sn.GetDevice(device); !ok {
			http.Error(w, "device not found", 404)
			return
		}
	}
	nI, _ := s.spec.GetNetworkIndex(network)
	nc := s.spec.Networks[nI].CensorForDevice(device)
	data, err := json.Marshal(nc)
	if err != nil {
		panic(err)
		return
	}
	w.WriteHeader(200)
	w.Write(data)
}

type PatchReifySpecRequest struct {
	ListenPort             int
	ListenPortSet          bool
	PublicKey              goal.Key
	PublicKeySet           bool
	PresharedKey           *goal.Key
	PresharedKeySet        bool
	PersistentKeepalive    time.Duration
	PersistentKeepaliveSet bool
}

func (s *Server) patchReifySpec(w http.ResponseWriter, r *http.Request) {
	network := r.PathValue("network")
	device := r.PathValue("device")
	if !s.verifyIdentity(w, r, network, device) {
		return
	}
	s.specLock.Lock()
	defer s.specLock.Unlock()
	{
		sn, ok := s.spec.GetNetwork(network)
		if !ok {
			http.Error(w, "network not found", 404)
			return
		}
		if _, ok = sn.GetDevice(device); !ok {
			http.Error(w, "device not found", 404)
			return
		}
	}
	var req PatchReifySpecRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, fmt.Sprintf("json decode failed: %s", err), 422)
	}
	nI, _ := s.spec.GetNetworkIndex(network)
	sndI, _ := s.spec.Networks[nI].GetDeviceIndex(device)
	if req.ListenPortSet {
		s.spec.Networks[nI].Devices[sndI].ListenPort = req.ListenPort
	}
	if req.PublicKeySet {
		s.spec.Networks[nI].Devices[sndI].PublicKey = req.PublicKey
	}
	if req.PresharedKeySet {
		s.spec.Networks[nI].Devices[sndI].PresharedKey = req.PresharedKey
	}
	if req.PersistentKeepaliveSet {
		s.spec.Networks[nI].Devices[sndI].PersistentKeepalive = req.PersistentKeepalive
	}
	w.WriteHeader(200)
	w.Write([]byte("done"))
	return
}

type PostReifyStatusRequest struct {
	Reified spec.NetworkCensored
}

func (s *Server) postReifyStatus(w http.ResponseWriter, r *http.Request) {
	network := r.PathValue("network")
	device := r.PathValue("device")
	if !s.verifyIdentity(w, r, network, device) {
		return
	}
	s.specLock.Lock()
	defer s.specLock.Unlock()
	{
		sn, ok := s.spec.GetNetwork(network)
		if !ok {
			http.Error(w, "network not found", 404)
			return
		}
		if _, ok = sn.GetDevice(device); !ok {
			http.Error(w, "device not found", 404)
			return
		}
	}
	var req PostReifyStatusRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, fmt.Sprintf("json decode failed: %s", err), 422)
	}
	nI, _ := s.spec.GetNetworkIndex(network)
	if !req.Reified.Equal(s.spec.Networks[nI].CensorForDevice(device)) {
		http.Error(w, "given network does not match mine", 422)
		return
	}

	if s.latest == nil {
		s.latest = map[string][]string{}
	}
	s.latest[network] = append(s.latest[network], device)
	w.WriteHeader(200)
	w.Write([]byte("done"))
	return
}
