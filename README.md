# Raft Consensus Algorithm Implementation

This project is an implementation of the Raft consensus algorithm, which allows a cluster of nodes to agree on the state of a distributed system.

## Getting Started

To create a multi-node cluster, you need to run each node in a separate terminal. Here's how to start a three-node cluster:

**Terminal 1: Leader Node**

```bash
go run main.go -nodeAddr 127.0.0.1:7001 -httpPort 8001
```

**Terminal 2: Follower Node**

```bash
go run main.go -nodeAddr 127.0.0.1:7002 -httpPort 8002 -leaderAddr 127.0.0.1:8001
```

**Terminal 3: Follower Node**

```bash
go run main.go -nodeAddr 127.0.0.1:7003 -httpPort 8003 -leaderAddr 127.0.0.1:8001
```

In the above commands:

* `nodeAddr` is the address of the node.
* `httpPort` is the HTTP port for the node.
* `leaderAddr` is the HTTP address of the leader node. This parameter is not needed for the leader node.

### Checking Node Status
A node can be in one of two states: leader (e.g. "goose") or follower (e.g. "duck"). You can check the status of a node by sending a GET request to its HTTP endpoint:

```bash
curl http://127.0.0.1:8001/
```

The response will indicate whether the node is the leader:

```bash
{"state":"goose"}
```

or a follower:

```bash
{"state":"duck"}
```

### Leader Election
If the leader node is killed (e.g., by using Control+C in its terminal), the remaining nodes will hold an election to choose a new leader. This is a key feature of the Raft consensus algorithm, which ensures that the cluster can continue to operate even if one node fails.

You can check the new leader by sending a GET request to the HTTP endpoint of another node:

```bash
curl http://127.0.0.1:8002/
{"state":"goose"}
```

### Scaling the Cluster
To add more nodes to the cluster, run the same command as for the follower nodes, but with a unique `nodeAddr` and `httpPort` for each new node.

### Elasticity
This implemetation supports dynamic cluster membership changes. This is acheived through the HTTP handlers `JoinHandler` and `LeaveHandler`.

**Adding Nodes:** The `JoinHandler` function handles a "join" request from a new node. It first adds the new node as a non-voter using `raft.AddNonvoter()`. Once the new node is added, it is promoted to a voter using `raft.AddVoter()`. This allows the cluster to expand dynamically as new nodes join.

**Removing Nodes:** The `LeaveHandler` function handles a "leave" request from a node that wants to leave the cluster. It removes the node from the cluster using `raft.RemoveServer()`. This allows the cluster to shrink dynamically as nodes leave.

### Troubleshooting
If you encounter any issues, please check the following:

Ensure all nodes are running and can communicate with each other.

Make sure the leader node is running when other nodes are started, so they can join the cluster.

If a node can't join the cluster, check the leader node's logs for any error messages.