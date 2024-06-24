package device

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/nyiyui/qrystal/coord"
	"go.uber.org/zap"
)

type ContinousClient struct {
	Client *Client
	latest bool
}

func (c *ContinousClient) Step() (latest, updated bool, err error) {
	var f func() (latest bool, err error)
	if c.latest {
		f = c.getLatest
		updated = false
	} else {
		f = c.Client.ReifySpec
		updated = true
	}
	var newLatest bool
	newLatest, err = f()
	if err != nil {
		return
	}
	latest = newLatest
	c.latest = newLatest
	return
}

func (c *ContinousClient) getLatest() (latest bool, err error) {
	req, err := http.NewRequest("GET", c.Client.baseURL.JoinPath(fmt.Sprintf("/v1/reify/%s/%s/latest", c.Client.network, c.Client.device)).String(), nil)
	if err != nil {
		panic(err)
	}
	c.Client.addAuthorizationHeader(req)
	resp, err := c.Client.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("get latest: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("get latest: %w", err)
	}
	if resp.StatusCode != 200 {
		return false, fmt.Errorf("get latest: %s: %s", resp.Status, data)
	}
	var respData coord.GetReifyLatestResponse
	err = json.Unmarshal(data, &respData)
	if err != nil {
		zap.S().Debugf("received body:\n%s", data)
		return false, fmt.Errorf("get latest: %w", err)
	}
	zap.S().Debugf("received body:\n%s", data)
	zap.S().Debugf("respData: %#v", respData)
	return respData.Latest, nil
}
