package rafthttp

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/derekschultz/raft-cluster/raftcluster"
	"github.com/hashicorp/raft"
)

const (
	Duck  = "duck"
	Goose = "goose"
)

// HttpServer represents the HTTP server for serving status requests
type HttpServer struct {
	Addr string
	Raft *raft.Raft
	Fsm  *raftcluster.NodeState
	mtx  sync.Mutex // Mutex to protect the node state
}

// ServeHTTP handles all HTTP requests.
// It locks the NodeState to prevent concurrent access, updates the node state based on the Raft state, and responds with the current node state.
func (s *HttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Mutex is used to ensure that only one goroutine can access the NodeState at a time.
	// This is necessary because the NodeState is shared between the Raft instance and the HTTP server.
	s.mtx.Lock()
	defer s.mtx.Unlock()

	// Update the node state based on the Raft state
	// If the node is the leader, it's a "goose". Otherwise, it's a "duck".
	if s.Raft.State() == raft.Leader {
		s.Fsm.State = Goose
	} else {
		s.Fsm.State = Duck
	}

	// Respond with the node state
	err := json.NewEncoder(w).Encode(s.Fsm)
	if err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// JoinHandler handles a "join" request from a new node.
// It adds the new node as a non-voter initially to allow it to catch up with the cluster state, and then promotes it to a voter.
func (s *HttpServer) JoinHandler(w http.ResponseWriter, r *http.Request) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	// Get the address of the new node from the request query parameters
	nodeAddr := r.URL.Query().Get("nodeAddr")
	if nodeAddr == "" {
		http.Error(w, "nodeAddr query parameter is required", http.StatusBadRequest)
		return
	}

	// Add the new node as a non-voter
	// A non-voter is a node that receives replicated log entries but does not participate in the quorum of the cluster.
	future := s.Raft.AddNonvoter(raft.ServerID(nodeAddr), raft.ServerAddress(nodeAddr), 0, 0)
	if err := future.Error(); err != nil {
		http.Error(w, "Failed to add non-voter: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Promote the new node to a voter
	// A voter is a node that participates in the quorum of the cluster.
	future = s.Raft.AddVoter(raft.ServerID(nodeAddr), raft.ServerAddress(nodeAddr), 0, 0)
	if err := future.Error(); err != nil {
		http.Error(w, "Failed to add voter: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Successfully added voter"))
}

// LeaveHandler handles a "leave" request from a node that wants to leave the cluster.
// It removes the node from the cluster.
func (s *HttpServer) LeaveHandler(w http.ResponseWriter, r *http.Request) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	// Get the address of the node from the request query parameters
	nodeAddr := r.URL.Query().Get("nodeAddr")
	if nodeAddr == "" {
		http.Error(w, "nodeAddr query parameter is required", http.StatusBadRequest)
		return
	}

	// Remove the node from the cluster
	future := s.Raft.RemoveServer(raft.ServerID(nodeAddr), 0, 0)
	if err := future.Error(); err != nil {
		http.Error(w, "Failed to remove server: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Successfully removed server"))
}
