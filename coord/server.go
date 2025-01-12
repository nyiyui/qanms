package coord

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"sync"

	"github.com/google/go-cmp/cmp"
	"github.com/nyiyui/qrystal/goal"
	"github.com/nyiyui/qrystal/spec"
	"github.com/nyiyui/qrystal/util"
	"go.uber.org/zap"
)

type TokenInfo struct {
	Identities [][2]string
}

type Server struct {
	mux      *http.ServeMux
	spec     spec.Spec
	specLock sync.RWMutex
	// latest lists which devices have applied the latest spec.
	// The key is the network name, and the value is the list of device names.
	latest     map[string][]string
	latestLock sync.RWMutex
	tokens     map[util.TokenHash]TokenInfo
}

func NewServer(spec spec.Spec, tokens map[util.TokenHash]TokenInfo) *Server {
	if tokens == nil {
		panic("coord.NewServer: tokens map must not be nil")
	}
	s := &Server{
		mux:    http.NewServeMux(),
		spec:   spec,
		latest: map[string][]string{},
		tokens: tokens,
	}
	s.setup()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) setup() {
	s.mux.HandleFunc("GET /v1/reify/{network}/{device}/latest", s.getReifyLatest)
	s.mux.HandleFunc("GET /v1/reify/{network}/{device}/spec", s.getReifySpec)
	s.mux.HandleFunc("PATCH /v1/reify/{network}/{device}/spec", s.patchReifySpec)
	s.mux.HandleFunc("POST /v1/reify/{network}/{device}/status", s.postReifyStatus)
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

type GetReifyLatestResponse struct {
	Latest bool
}

func (s *Server) getReifyLatest(w http.ResponseWriter, r *http.Request) {
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
	s.latestLock.RLock()
	defer s.latestLock.RUnlock()
	var response string
	if slices.Contains(s.latest[network], device) {
		response = `{"Latest":true}`
	} else {
		response = `{"Latest":false}`
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte(response))
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
	PersistentKeepalive    goal.Duration
	PersistentKeepaliveSet bool

	// Accessible is the list of devices accessible without forwarding.
	Accessible    []string
	AccessibleSet bool
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
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", 500)
		return
	}
	var req PatchReifySpecRequest
	err = json.Unmarshal(data, &req)
	if err != nil {
		http.Error(w, fmt.Sprintf("json decode failed: %s\n%s", err, data), 400)
		return
	}
	newSpec := s.spec.Clone()
	nI, _ := newSpec.GetNetworkIndex(network)
	sndI, _ := newSpec.Networks[nI].GetDeviceIndex(device)
	if req.ListenPortSet {
		newSpec.Networks[nI].Devices[sndI].ListenPort = req.ListenPort
	}
	if req.PublicKeySet {
		newSpec.Networks[nI].Devices[sndI].PublicKey = req.PublicKey
	}
	if req.PresharedKeySet {
		newSpec.Networks[nI].Devices[sndI].PresharedKey = req.PresharedKey
	}
	if req.PersistentKeepaliveSet {
		newSpec.Networks[nI].Devices[sndI].PersistentKeepalive = req.PersistentKeepalive
	}
	if req.AccessibleSet {
		for _, name := range req.Accessible {
			if _, ok := newSpec.Networks[nI].GetDevice(name); !ok {
				http.Error(w, fmt.Sprintf("Accessible contains nonexistent device name: %s/%s", network, name), 400)
				return
			}
		}
		zap.S().Debugf("setting Accessible to %v", req.Accessible)
		newSpec.Networks[nI].Devices[sndI].Accessible = req.Accessible
	}
	s.latestLock.Lock()
	defer s.latestLock.Unlock()
	s.updateSpecNoLock(newSpec)
	data, err = json.Marshal(s.spec.Networks[nI].Devices[sndI])
	if err != nil {
		panic(err)
	}
	zap.S().Infof("patched %s/%s:\n%s", network, device, data)
	w.WriteHeader(204)
	return
}

type PostReifyStatusRequest struct {
	Reified spec.NetworkCensored
}

type PostReifyStatusResponse struct {
	Latest bool
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
		http.Error(w, fmt.Sprintf("request data read or json decode failed: %s", err), 400)
	}
	nI, ok := s.spec.GetNetworkIndex(network)
	if !ok {
		http.Error(w, "invalid request data", 422)
		return
	}
	if !req.Reified.Equal(s.spec.Networks[nI].CensorForDevice(device)) {
		zap.S().Infof("given network does not match mine (mine minus given):\n%s", cmp.Diff(req.Reified, s.spec.Networks[nI].CensorForDevice(device)))
		data, _ := json.Marshal(req.Reified)
		zap.S().Infof("given network:\n%s", data)
		data, _ = json.Marshal(s.spec.Networks[nI].CensorForDevice(device))
		zap.S().Infof("my network:\n%s", data)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		err := json.NewEncoder(w).Encode(PostReifyStatusResponse{false})
		if err != nil {
			zap.S().Error("json encode and HTTP write of PostReifyStatusResponse failed: %s", err)
		}
		return
	}

	s.latestLock.Lock()
	defer s.latestLock.Unlock()
	if s.latest == nil {
		s.latest = map[string][]string{}
	}
	s.latest[network] = append(s.latest[network], device)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	err = json.NewEncoder(w).Encode(PostReifyStatusResponse{true})
	if err != nil {
		zap.S().Error("json encode and HTTP write of PostReifyStatusResponse failed: %s", err)
	}
	return
}
