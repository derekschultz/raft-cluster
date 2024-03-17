package raftcluster

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/raft"
	boltdbraft "github.com/hashicorp/raft-boltdb"
)

// SetupRaft sets up the Raft instance for the current node.
// It creates the Raft configuration, communication, and storage, and joins the cluster.
func SetupRaft(nodeAddr string, cluster []string, fsm *NodeState) (*raft.Raft, error) {
	// Set up Raft configuration
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(nodeAddr)

	// Set up Raft communication
	addr, err := net.ResolveTCPAddr("tcp", nodeAddr)
	if err != nil {
		return nil, err
	}
	transport, err := raft.NewTCPTransport(nodeAddr, addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return nil, err
	}

	// Set up the Raft storage
	raftDir := filepath.Join("raft", nodeAddr)
	// Create the directory if it does not exist
	err = os.MkdirAll(raftDir, 0755)
	if err != nil {
		return nil, err
	}

	logStore, err := boltdbraft.NewBoltStore(filepath.Join(raftDir, "raft-log.bolt"))
	if err != nil {
		return nil, err
	}
	stableStore, err := boltdbraft.NewBoltStore(filepath.Join(raftDir, "raft-stable.bolt"))
	if err != nil {
		return nil, err
	}
	snapshotStore, err := raft.NewFileSnapshotStore(raftDir, 2, os.Stderr)
	if err != nil {
		return nil, err
	}

	// Create the Raft system
	raftInstance, err := raft.NewRaft(config, fsm, logStore, stableStore, snapshotStore, transport)
	if err != nil {
		return nil, err
	}

	// Add the current node to the cluster
	err = addNodeToCluster(raftInstance, config, transport, cluster)
	if err != nil {
		return nil, err
	}

	return raftInstance, nil
}

// addNodeToCluster adds the current node to the cluster.
func addNodeToCluster(raftInstance *raft.Raft, config *raft.Config, transport *raft.NetworkTransport, cluster []string) error {
	if len(cluster) == 0 {
		// Get the current Raft configuration
		future := raftInstance.GetConfiguration()
		if err := future.Error(); err != nil {
			return fmt.Errorf("failed to get configuration: %v", err)
		}

		// Check if the cluster has already been bootstrapped
		if len(future.Configuration().Servers) == 0 {
			// The cluster has not been bootstrapped, bootstrap it
			bootstrapConfig := raft.Configuration{
				Servers: []raft.Server{
					{
						ID:      config.LocalID,
						Address: transport.LocalAddr(),
					},
				},
			}
			configFuture := raftInstance.BootstrapCluster(bootstrapConfig)
			if err := configFuture.Error(); err != nil {
				return fmt.Errorf("failed to bootstrap cluster: %v", err)
			}
		}
	} else {
		// Check if the current node is the leader
		if raftInstance.State() == raft.Leader {
			log.Println("Promoting to voter:", transport.LocalAddr())
			future := raftInstance.AddVoter(config.LocalID, raft.ServerAddress(transport.LocalAddr()), 0, 0)
			if err := future.Error(); err != nil {
				return fmt.Errorf("failed to add voter: %v", err)
			}
			log.Println("Promoted to voter:", transport.LocalAddr())
		} else {
			log.Println("Current node is not the leader. Skipping promoting to voter.")
		}
	}

	return nil
}
