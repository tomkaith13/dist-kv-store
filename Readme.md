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
goos: darwin
goarch: amd64
pkg: github.com/tomkaith13/dist-kv-store/internal/service
cpu: Intel(R) Core(TM) i7-6820HQ CPU @ 2.70GHz
BenchmarkNoRaftWithOneNodesSetAndGet-8   	       1	4510991898 ns/op	728895336 B/op	 9701978 allocs/op
PASS
```

With Raft cluster setup:
Using k6
- Install [k6](https://grafana.com/docs/k6/latest/set-up/install-k6/) 
- kick start all nodes in this order:
  -  `make run1` , `make run2` and finally `make run3`
- run `k6 run perf.js`

The results are:
```bash
➜  dist-key-val git:(main) ✗ k6 run perf.js

          /\      |‾‾| /‾‾/   /‾‾/
     /\  /  \     |  |/  /   /  /
    /  \/    \    |     (   /   ‾‾\
   /          \   |  |\  \ |  (‾)  |
  / __________ \  |__| \__\ \_____/ .io

     execution: local
        script: perf.js
        output: -

     scenarios: (100.00%) 1 scenario, 1 max VUs, 40s max duration (incl. graceful stop):
              * default: 1 looping VUs for 10s (gracefulStop: 30s)


     ✓ post status was 201
     ✓ get status was 200

     checks.........................: 100.00% ✓ 20       ✗ 0
     data_received..................: 2.8 kB  276 B/s
     data_sent......................: 2.8 kB  275 B/s
     dkv_get_key....................: avg=538.69µs min=396µs med=518µs    max=868µs  p(90)=688µs   p(95)=777.99µs
     dkv_set_key....................: avg=1.27ms   min=664µs med=782.5µs  max=4.22ms p(90)=2.48ms  p(95)=3.35ms
     http_req_blocked...............: avg=124.4µs  min=3µs   med=5µs      max=2.03ms p(90)=46.6µs  p(95)=450.2µs
     http_req_connecting............: avg=46.8µs   min=0s    med=0s       max=683µs  p(90)=25.3µs  p(95)=274.5µs
     http_req_duration..............: avg=908.45µs min=396µs med=686µs    max=4.22ms p(90)=1.15ms  p(95)=2.38ms
       { expected_response:true }...: avg=908.45µs min=396µs med=686µs    max=4.22ms p(90)=1.15ms  p(95)=2.38ms
     http_req_failed................: 0.00%   ✓ 0        ✗ 20
     http_req_receiving.............: avg=77.09µs  min=41µs  med=55µs     max=204µs  p(90)=150.3µs p(95)=172.65µs
     http_req_sending...............: avg=199.9µs  min=17µs  med=32µs     max=3.08ms p(90)=127µs   p(95)=368.95µs
     http_req_tls_handshaking.......: avg=0s       min=0s    med=0s       max=0s     p(90)=0s      p(95)=0s
     http_req_waiting...............: avg=631.44µs min=330µs med=601.49µs max=1.86ms p(90)=909.7µs p(95)=1.01ms
     http_reqs......................: 20      1.992157/s
     iteration_duration.............: avg=1s       min=1s    med=1s       max=1s     p(90)=1s      p(95)=1s
     iterations.....................: 10      0.996079/s
     vus............................: 1       min=1      max=1
     vus_max........................: 1       min=1      max=1


running (10.0s), 0/1 VUs, 10 complete and 0 interrupted iterations
default ✓ [======================================] 1 VUs  10s
```

The results can be found in the trends with names `dkv_get_key` and `dkv_set_key`.

### Improvements
- As an improvement we could add support for leader forwarding. Say an old leader from a prev term gets a SET /key request. A better UX would be to forward that request over to the new leader. Currently that functionality is absent.
-limit the key and val size to ensure the snapshotting process is quick and same goes with restore.
- Figure out how to setup the cluster and test without using tools like k6
    - right now, I keep getting connection refused when i try to add a second node in tests.

