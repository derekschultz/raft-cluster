package raftcluster

import (
	"encoding/json"
	"io"

	"github.com/hashicorp/raft"
)

// NodeState represents the state of a node (Duck or Goose)
type NodeState struct {
	State string `json:"state"`
}

// Apply applies a Raft log entry to the NodeState
func (n *NodeState) Apply(log *raft.Log) interface{} {
	n.State = string(log.Data)
	return nil
}

// Restore restores the NodeState from a snapshot
func (n *NodeState) Restore(rc io.ReadCloser) error {
	defer rc.Close()
	return json.NewDecoder(rc).Decode(n)
}

// Persist saves the NodeState to a writer
func (n *NodeState) Persist(sink raft.SnapshotSink) error {
	err := func() error {
		// Encode data
		b, err := json.Marshal(n)
		if err != nil {
			return err
		}

		// Write data to sink
		if _, err := sink.Write(b); err != nil {
			return err
		}

		return nil
	}()

	// If there was an error, cancel the sink and return the error
	if err != nil {
		sink.Cancel()
		return err
	}

	return nil
}

// Snapshot returns a snapshot of the NodeState
func (n *NodeState) Snapshot() (raft.FSMSnapshot, error) {
	return n, nil
}

// Release is called when raft is finished with the snapshot.
func (n *NodeState) Release() {}
