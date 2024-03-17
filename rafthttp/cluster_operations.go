package rafthttp

import (
	"fmt"
	"io"
	"net/http"
)

// JoinCluster sends a "join" request to the leader node to join the cluster.
func JoinCluster(nodeAddr string, leaderAddr string) error {
	resp, err := http.Get(fmt.Sprintf("http://%s/join?nodeAddr=%s", leaderAddr, nodeAddr))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to join cluster: %s", body)
	}

	return nil
}
