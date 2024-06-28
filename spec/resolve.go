package spec

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"sync"

	"go.uber.org/zap"
)

type EndpointScorer func(endpoint string) (int, error)

// PingCommandScorer gives a score of 0 for all reachable (see ping command below) hosts, and an error for unreachable ones.
// The command used is: ping -c 1 -w 4 -- <endpoint>
func PingCommandScorer(endpoint string) (int, error) {
	// NOTE: ping count, ping deadlines are arbitrarily set
	udpAddr, err := net.ResolveUDPAddr("udp", endpoint)
	if err != nil {
		return 0, err
	}
	cmd := exec.Command("ping", "-c", "1", "-w", "4", "--", udpAddr.IP.String())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return 0, err
	}
	return 0, nil
}

var ErrAllEndpointsBad = errors.New("spec: all endpoints are bad")

// ChooseEndpoint chooses an endpoint and forwarder using the score function provided.
// If the score returned is tied, the first endpoint (in NetworkDeviceCensored.Endpoints) is chosen.
// The score function is run in separate goroutines for each endpoint.
// If all scorers return an error, a forwarder is chosen instead.
func (ndc *NetworkDeviceCensored) ChooseEndpoint(score EndpointScorer) error {
	// TODO: if all endpoints fail, randomly choose a forwarder (look in ForwardsFor)
	scores := make([]int, len(ndc.Endpoints))
	errs := make([]error, len(ndc.Endpoints))
	var wg sync.WaitGroup
	for i, endpoint := range ndc.Endpoints {
		wg.Add(1)
		go func(i int, endpoint string) {
			scores[i], errs[i] = score(endpoint)
			zap.S().Debugf("%d: score=%d, err=%s", i, scores[i], errs[i])
			wg.Done()
		}(i, endpoint)
	}
	wg.Wait()
	ok := false
	for _, err := range errs {
		if err == nil {
			ok = true
		}
	}
	if !ok {
		// all scorers returned an error - try forwarders
		return fmt.Errorf("all scorers returned an error: %w", ErrAllEndpointsBad)
	}
	maxI := -1
	maxScore := -1
	for i, score := range scores {
		if errs[i] != nil {
			continue
		}
		if score > maxScore {
			maxI, maxScore = i, score
		}
	}
	if maxI == -1 {
		panic("unreachable")
	}
	ndc.EndpointChosenIndex = maxI
	ndc.ForwarderAndEndpointChosen = true
	return nil
}
