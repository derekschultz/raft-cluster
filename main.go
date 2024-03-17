package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/derekschultz/raft-cluster/raftcluster"
	"github.com/derekschultz/raft-cluster/rafthttp"
)

func main() {
	// Define the command line arguments
	nodeAddr := flag.String("nodeAddr", "", "address of this node for Raft communication")
	httpPort := flag.String("httpPort", "", "port of this node for HTTP communication")
	leaderAddr := flag.String("leaderAddr", "", "address of the leader node for HTTP communication") // Optional

	// Pase the command line arguments
	flag.Parse()

	// Check that the required command line arguments were provided
	if *nodeAddr == "" || *httpPort == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Set up Raft
	fsm := &raftcluster.NodeState{}
	var cluster []string
	if *leaderAddr != "" {
		cluster = append(cluster, *leaderAddr)
	}
	raftInstance, err := raftcluster.SetupRaft(*nodeAddr, cluster, fsm)
	if err != nil {
		log.Fatal(err)
	}

	// If this node is not the first node in the cluster, join the cluster
	if *leaderAddr != "" {
		err := rafthttp.JoinCluster(*nodeAddr, *leaderAddr)
		if err != nil {
			log.Fatalf("Failed to join cluster: %v", err)
		}
	}

	// Create the HTTP server
	server := &rafthttp.HttpServer{
		Addr: *httpPort,
		Raft: raftInstance,
		Fsm:  fsm,
	}

	http.Handle("/", server)
	http.HandleFunc("/join", server.JoinHandler)
	http.HandleFunc("/leave", server.LeaveHandler)
	log.Fatal(http.ListenAndServe(":"+server.Addr, nil))
}
