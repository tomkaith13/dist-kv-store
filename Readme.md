# Distributed Key-Value Store with Raft

## Problem
Create a key-value store that runs across a cluster. The data is replicated using [Raft(https://raft.github.io/)]
to achieve data replication in a strongly consistent fashion as long as there is quorum established.

Suppose there are $n$ nodes in this cluster that maintain the kv store, we define a **quorum is established** if the following is true:

$(n/2) + 1\;nodes\,are\,available$

As long as the quorum stays at this level, the cluster can be consider available and fault-tolerant.

Nodes can be removed and added back to the cluster and would be eventually consistent thanks to the snapshotting and restore mechanism of Raft protocol.

## How to run this code
This repo uses `.env*` files to initialize the cluster nodes.
So to run a cluster, the following steps need to be taken:
- Run `make run-test` to verify if your repo works locally.
- Copy and edit the 3 env files (`.env1`, `.env2` and `.env3`) to create 3 separate initialization configurations
- ***Optional***: You can create more `.env` files if you want to add more nodes to the cluster and edit the `Makefile` to add more rules. The current `Makefile` only has rules for 3 nodes.
- Run `make run1` to kickstart the first node. This automatically becomes the `leader` node.
- Run `make run2` to add the second node. Any nodes after the leader would join as a follower and the FSM state would be replicated to the follower.
- Run `make run3` 
- Call `POST leaderaddr/key` with body to store kv pair
- Call `GET nodeaddr/key/{key}` to fetch the pair from any node in the cluster

## Fault Tolerance
Feel free to kill any node in the cluster and as long as the **quorum condition** is met, the cluster should still be available.

Killing off the `leader` would trigger a *leader-election* and any new `POSTs` would need to be done on the new leader. Request forwarding to the new leader is not covered in this example.

Adding a node back into the mix, should trigger raft to kick in and re-populate the internal store.
