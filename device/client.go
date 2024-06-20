package device

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/nyiyui/qrystal/coord"
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

	token      util.Token
	network    string
	device     string
	privateKey goal.Key
}

func NewClient(baseURL string, token util.Token, network, device string, privateKey goal.Key) (*Client, error) {
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
	return &Client{
		client:        new(http.Client),
		baseURL:       baseURL2,
		wgClient:      wgClient,
		netlinkHandle: netlinkHandle,
		token:         token,
		network:       network,
		device:        device,
		privateKey:    privateKey,
	}, nil
}

func (c *Client) ReifySpec() error {
	resp, err := c.client.Get(c.baseURL.JoinPath(fmt.Sprintf("/v1/reify/%s/%s/spec", c.network, c.device)).String())
	if err != nil {
		return fmt.Errorf("get spec: %w", err)
	}
	defer resp.Body.Close()
	var nc spec.NetworkCensored
	err = json.NewDecoder(resp.Body).Decode(&nc)
	if err != nil {
		return fmt.Errorf("get spec: %w", err)
	}
	data, _ := json.Marshal(nc)
	zap.S().Debugf("received spec:\n%s", data)

	ndc, _ := nc.GetDevice(c.device)
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
			return fmt.Errorf("patch spec: %w", err)
		}
		ndc.PublicKey = goal.Key(wgtypes.Key(c.privateKey).PublicKey())
		zap.S().Debug("patched spec public key.")
	}

	// === choose endpoints ===
	if !ndc.EndpointChosen {
		zap.S().Debug("choosing endpoint…")
		ndcI, _ := nc.GetDeviceIndex(c.device)
		err = nc.Devices[ndcI].ChooseEndpoint(spec.PingCommandScorer)
		if err != nil {
			return fmt.Errorf("choose endpoint: %w", err)
		}
		zap.S().Debugf("endpoint %s chosen.", nc.Devices[ndcI].Endpoints[nc.Devices[ndcI].EndpointChosenIndex])
	}

	// === apply spec ===
	zap.S().Debug("compiling spec…")
	sc := spec.SpecCensored{Networks: []spec.NetworkCensored{nc}}
	gm, err := sc.CompileMachine(c.device)
	if err != nil {
		return fmt.Errorf("compile spec: %w", err)
	}
	gm.Interfaces[0].PrivateKey = goal.Key(c.privateKey)
	data, _ = json.Marshal(gm)
	zap.S().Debugf("compiled spec:\n%s", data)
	zap.S().Debug("applying machine…")
	err = goal.ApplyMachineDiff(c.Machine, gm, goal.DiffMachine(&c.Machine, &gm), c.wgClient, c.netlinkHandle)
	if err != nil {
		return fmt.Errorf("apply spec: %w", err)
	}
	zap.S().Debug("applied machine.")

	// === post status ===
	zap.S().Debug("posting status…")
	data, err = json.Marshal(coord.PostReifyStatusRequest{
		Reified: nc,
	})
	if err != nil {
		panic(fmt.Sprintf("json marshal: %s", err))
	}
	resp, err = c.client.Post(c.baseURL.JoinPath(fmt.Sprintf("/v1/reify/%s/%s/status", c.network, c.device)).String(), "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("post status: %w", err)
	}
	zap.S().Debug("posted status.")
	return nil
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
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("post status: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s: %s", resp.Status, data)
	}
	return nil
}
