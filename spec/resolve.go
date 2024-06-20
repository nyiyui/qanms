package spec

import (
	"fmt"
	"os/exec"
	"sync"
)

type EndpointScorer func(endpoint string) (int, error)

// PingCommandScorer gives a score of 0 for all reachable (see ping command below) hosts, and an error for unreachable ones.
// The command used is: ping -c 1 -w 4 -- <endpoint>
func PingCommandScorer(endpoint string) (int, error) {
	// NOTE: ping count, ping deadlines are arbitrarily set
	cmd := exec.Command("ping", "-c", "1", "-w", "4", "--", endpoint)
	err := cmd.Run()
	if err != nil {
		return 0, err
	}
	return 0, nil
}

// ChooseEndpoint chooses an endpoint using the score function provided.
// If the score returned is tied, the first endpoint (in NetworkDeviceCensored.Endpoints) is chosen.
// The score function is run in separate goroutines for each endpoint.
func (ndc *NetworkDeviceCensored) ChooseEndpoint(score EndpointScorer) error {
	scores := make([]int, len(ndc.Endpoints))
	errs := make([]error, len(ndc.Endpoints))
	var wg sync.WaitGroup
	for i, endpoint := range ndc.Endpoints {
		wg.Add(1)
		go func(i int, endpoint string) {
			scores[i], errs[i] = score(endpoint)
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
		return fmt.Errorf("all scorers returned an error. first error: %w", errs[0])
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
	ndc.EndpointChosen = true
	return nil
}
