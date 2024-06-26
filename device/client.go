package device

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"

	"github.com/nyiyui/qrystal/coord"
	"github.com/nyiyui/qrystal/dns"
	"github.com/nyiyui/qrystal/goal"
	"github.com/nyiyui/qrystal/spec"
	"github.com/nyiyui/qrystal/util"
	"github.com/vishvananda/netlink"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type Client struct {
	client        *http.Client
	baseURL       *url.URL
	Machine       goal.Machine
	wgClient      *wgctrl.Client
	netlinkHandle *netlink.Handle
	dns           dns.Client
	dnsLock       sync.Mutex

	spec       spec.SpecCensored
	token      util.Token
	network    string
	device     string
	privateKey goal.Key
}

func NewClient(httpClient *http.Client, baseURL string, token util.Token, network, device string, privateKey goal.Key) (*Client, error) {
	baseURL2, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	wgClient, err := wgctrl.New()
	if err != nil {
		return nil, err
	}
	netlinkHandle, err := netlink.NewHandle()
	if err != nil {
		return nil, err
	}
	if httpClient == nil {
		httpClient = new(http.Client)
	}
	return &Client{
		client:        httpClient,
		baseURL:       baseURL2,
		wgClient:      wgClient,
		netlinkHandle: netlinkHandle,
		token:         token,
		network:       network,
		device:        device,
		privateKey:    privateKey,
	}, nil
}

func (c *Client) SetDNSClient(client dns.Client) {
	c.dnsLock.Lock()
	defer c.dnsLock.Unlock()
	c.dns = client
}

func (c *Client) updateDNS() error {
	c.dnsLock.Lock()
	defer c.dnsLock.Unlock()
	if c.dns != nil {
		zap.S().Debug("updating DNS server…")
		err := c.dns.UpdateSpec(c.spec)
		if err != nil {
			return fmt.Errorf("update DNS server: %w", err)
		}
		zap.S().Debug("done updating DNS server.")
	}
	return nil
}

func (c *Client) addAuthorizationHeader(r *http.Request) {
	r.Header.Set("Authorization", "QrystalCoordIdentityToken "+c.token.String())
}

func (c *Client) ReifySpec() (latest bool, err error) {
	path := c.baseURL.JoinPath(fmt.Sprintf("/v1/reify/%s/%s/spec", c.network, c.device)).String()
	zap.S().Debugf("path: %s", path)
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		panic(err)
	}
	c.addAuthorizationHeader(req)
	resp, err := c.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("get spec: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("get spec: %w", err)
	}
	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		return false, fmt.Errorf("get spec: %s: %s", resp.Status, data)
	}
	var nc spec.NetworkCensored
	err = json.Unmarshal(data, &nc)
	if err != nil {
		zap.S().Debugf("received body:\n%s", data)
		return false, fmt.Errorf("get spec: %w", err)
	}
	data, _ = json.Marshal(nc)
	zap.S().Debugf("received spec:\n%s", data)
	zap.S().Debugf("c.device = %s", c.device)

	ndcI, ok := nc.GetDeviceIndex(c.device)
	if !ok {
		panic("unreachable")
	}
	zap.S().Debugf("ndcI = %s", ndcI)
	ndc := &nc.Devices[ndcI]
	// === generate private keys ===
	if c.privateKey == (goal.Key{}) {
		zap.S().Debug("generating private keys…")
		privateKey, err := wgtypes.GeneratePrivateKey()
		if err != nil {
			panic(fmt.Sprintf("generate private key: %s", err))
		}
		c.privateKey = goal.Key(privateKey)
		zap.S().Debugf("generated key pair:\nprivate key: %s\npublic key: %s", privateKey, privateKey.PublicKey())
	}
	// === update spec's public keys ===
	if wgtypes.Key(ndc.PublicKey) != wgtypes.Key(c.privateKey).PublicKey() {
		zap.S().Debug("public key set in spec mismatch, patching spec…")
		err = c.patchSpec(coord.PatchReifySpecRequest{
			PublicKey:    goal.Key(wgtypes.Key(c.privateKey).PublicKey()),
			PublicKeySet: true,
		})
		if err != nil {
			return false, fmt.Errorf("patch spec: %w", err)
		}
		ndc.PublicKey = goal.Key(wgtypes.Key(c.privateKey).PublicKey())
		zap.S().Debug("patched spec public key.")
	}

	// === choose endpoints ===
	for i, ndc := range nc.Devices {
		zap.S().Debugf("i=%d ndcI=%d ndc.Name=%d", i, ndcI, ndc.Name)
		if i == ndcI {
			continue
		}
		if !ndc.EndpointChosen {
			zap.S().Debugf("%s/%s: choosing endpoint…", c.network, ndc.Name)
			//err = (&nc.Devices[i]).ChooseEndpoint(spec.PingCommandScorer)
			err = (&nc.Devices[i]).ChooseEndpoint(func(string) (int, error) { return 0, nil })
			if err != nil {
				return false, fmt.Errorf("choose endpoint for %s/%s: %w", c.network, ndc.Name, err)
			}
			if !nc.Devices[i].EndpointChosen {
				panic("unreachable")
			}
			zap.S().Debugf("%s/%s: endpoint %d (%s) chosen.", c.network, ndc.Name, ndc.EndpointChosenIndex, ndc.Endpoints[ndc.EndpointChosenIndex])
		}
	}

	data, _ = json.MarshalIndent(ndc, "", "  ")
	zap.S().Debugf("ndc:\n%s", data)
	data, _ = json.MarshalIndent(nc, "", "  ")
	zap.S().Debugf("nc:\n%s", data)
	c.spec = spec.SpecCensored{Networks: []spec.NetworkCensored{nc}}

	// === update DNS ===
	err = c.updateDNS()
	if err != nil {
		return false, fmt.Errorf("update DNS server: %w", err)
	}

	// === apply spec ===
	zap.S().Debug("compiling spec…")
	gm, err := c.spec.CompileMachine(c.device, true)
	if err != nil {
		return false, fmt.Errorf("compile spec: %w", err)
	}
	gm.Interfaces[0].PrivateKey = goal.Key(c.privateKey)
	data, _ = json.Marshal(gm)
	zap.S().Debugf("compiled spec:\n%s", data)
	zap.S().Debug("applying machine…")
	err = goal.ApplyMachineDiff(c.Machine, gm, goal.DiffMachine(&c.Machine, &gm), c.wgClient, c.netlinkHandle)
	if err != nil {
		return false, fmt.Errorf("apply spec: %w", err)
	}
	c.Machine = gm
	zap.S().Debug("applied machine.")

	// === post status ===
	zap.S().Debug("posting status…")
	data, err = json.Marshal(coord.PostReifyStatusRequest{
		Reified: nc,
	})
	if err != nil {
		panic(fmt.Sprintf("json marshal: %s", err))
	}
	req, err = http.NewRequest("POST", c.baseURL.JoinPath(fmt.Sprintf("/v1/reify/%s/%s/status", c.network, c.device)).String(), bytes.NewBuffer(data))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	c.addAuthorizationHeader(req)
	resp, err = c.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("post status: %w", err)
	}
	data, _ = io.ReadAll(resp.Body)
	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		return false, fmt.Errorf("post status: %s: %s", resp.Status, data)
	}
	var respData coord.PostReifyStatusResponse
	err = json.Unmarshal(data, &respData)
	if err != nil {
		return false, fmt.Errorf("json decode: %w", err)
	}
	return respData.Latest, nil
}

func (c *Client) patchSpec(body coord.PatchReifySpecRequest) error {
	data, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest("PATCH", c.baseURL.JoinPath(fmt.Sprintf("/v1/reify/%s/%s/spec", c.network, c.device)).String(), bytes.NewBuffer(data))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	c.addAuthorizationHeader(req)
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("post status: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s: %s", resp.Status, data)
	}
	return nil
}
