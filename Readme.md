# Distributed Key-Value Store with Raft

## Problem
Create a key-value store that runs across a cluster. The data is replicated using [Raft](https://raft.github.io/)
to achieve data replication in a strongly consistent fashion as long as there is quorum established.

Suppose there are $n$ nodes in this cluster that maintain the kv store, we define a **quorum is established** if the following is true:

$(n/2) + 1\ nodes\ are\ available$

As long as the quorum stays at this level, the cluster can be consider available and fault-tolerant.

Nodes can be removed and added back to the cluster and would be eventually consistent thanks to the snapshotting and restore mechanism of Raft protocol.

## How to run this code
This repo uses `.env*` files to initialize the cluster nodes.
So to run a cluster, the following steps need to be done:
- Install [Go v1.23](https://go.dev/doc/install)
- Run `make run-test` to verify if your repo works locally.
- Copy and edit the 3 env files (`.env1`, `.env2` and `.env3`) to create 3 separate initialization configurations
- ***Optional***: You can create more `.env` files if you want to add more nodes to the cluster and edit the `Makefile` to add more rules. The current `Makefile` only has rules for 3 nodes.
- Run `make run1` to kickstart the first node. This automatically becomes the `leader` node.
- Run `make run2` to add the second node. Any nodes after the leader would join as a follower and the FSM state would be replicated to the follower.
- Run `make run3` 
- Call `POST leaderaddr/key` with body to store kv pair
- Call `GET nodeaddr/key/{key}` to fetch the pair from any node in the cluster

## Configuration 
This section explains the configs found in the env files

```bash
# server configs
SERVER_ADDRESS=localhost:8889 ---> This is the address used to make the GET/POST/DEL with keys

# service kv configs
SERVICE_KEY_MAX_LEN=100 ---> We limit the size of the keys using this config. If larger, we get a 400
SERVICE_VAL_MAX_LEN=200 ---> We limit the size of the vals the same way
SERVICE_MAX_MAP_SIZE=1000 --> This is how we keep track of the upper limit of the size of the map. We get a 400 if this is exceeded as well

# service raft configs
SERVICE_RAFT_LEADER=false -------------------> this is used to indicate if the node (at setup time) is a leader or follower 
SERVICE_RAFT_STORE_DIR="./node2" ------------> log store location for raft
SERVICE_RAFT_ADDR=localhost:21002 -----------> raft addr
SERVICE_RAFT_NODE_ID=node2-------------------> raft node id
SERVICE_RAFT_JOIN_ADDR=localhost:8888--------> if this is a follower node, we need to register with the leader and this addr is used

PS: Service also has a `debug` config which is used in tests to run without raft. 
```
## Fault Tolerance
Feel free to kill any node in the cluster and as long as the **quorum condition** is met, the cluster should still be available.

Killing off the `leader` would trigger a *leader-election* and any new `POSTs` would need to be done on the new leader. Request forwarding to the new leader is not covered in this example.

Adding a node back into the mix, should trigger raft to kick in and re-populate the internal store.

## Benchmarks
Without Raft cluster setup and on 100k map size, we get:
```bash
1000000000	         0.0007660 ns/op	       0 B/op	       0 allocs/op
```

With Raft cluster setup:
Can be done using k6 scripts once the setup is done manually.

### Improvements
- As an improvement we could add support for leader forwarding. Say an old leader from a prev term gets a SET /key request. A better UX would be to forward that request over to the new leader. Currently that functionality is absent.
-limit the key and val size to ensure the snapshotting process is quick and same goes with restore.
- Figure out how to setup the cluster and test without using tools like k6
    - right now, I keep getting connection refused when i try to add a second node in tests.

