package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/hashicorp/raft"
	"github.com/rs/zerolog"
)

type DKVService struct {
	logger        zerolog.Logger
	ServiceConfig Config
	mu            sync.Mutex
	kvmap         map[string]string

	// raft FSM
	raft *raft.Raft
}

type Config struct {
	KeyMaxLen  int `envconfig:"KEY_MAX_LEN" default:"100"`
	ValMaxLen  int `envconfig:"VAL_MAX_LEN" default:"200"`
	MaxMapSize int `envconfig:"MAX_MAP_SIZE" default:"1000"`

	RaftNodeID   string        `envconfig:"RAFT_NODE_ID" required:"true"`
	RaftAddr     string        `envconfig:"RAFT_ADDR" required:"true"`
	RaftStoreDir string        `envconfig:"RAFT_STORE_DIR" required:"true"`
	RaftTimeout  time.Duration `envconfig:"RAFT_TIMEOUT" default:"20s"`

	Debug        bool
	RaftLeader   bool   `envconfig:"RAFT_LEADER" required:"true"`
	RaftJoinAddr string `envconfig:"RAFT_JOIN_ADDR"`
}

func New(logger zerolog.Logger, config Config) *DKVService {
	service := &DKVService{
		logger:        logger,
		ServiceConfig: config,
	}

	service.kvmap = make(map[string]string)
	service.PrintConfigs()
	if !config.Debug {
		service.initRaft()
	}
	return service
}

func (s *DKVService) initRaft() {
	// create store dir
	if err := os.MkdirAll(s.ServiceConfig.RaftStoreDir, 0700); err != nil {
		s.logger.Fatal().Msg("Unable to create local raft store directory")
	}
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(s.ServiceConfig.RaftNodeID)

	addr, err := net.ResolveTCPAddr("tcp", s.ServiceConfig.RaftAddr)
	if err != nil {
		s.logger.Fatal().Msgf("Error getting a tcp endpoiint for raft. Err: %q", err)
		return
	}

	transport, err := raft.NewTCPTransport(s.ServiceConfig.RaftAddr, addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		s.logger.Fatal().Msgf("Error getting a tcp transport layer for raft. Err: %q", err)
		return
	}

	snapshots, err := raft.NewFileSnapshotStore(s.ServiceConfig.RaftStoreDir, 2, s.logger)
	if err != nil {
		s.logger.Fatal().Msgf("Error creating a snapshot store for raft. Err: %q", err)
		return
	}
	logStore := raft.NewInmemStore()
	stableStore := raft.NewInmemStore()

	// Instantiate the Raft systems.
	s.raft, err = raft.NewRaft(config, (*DKVService)(s), logStore, stableStore, snapshots, transport)
	if err != nil {
		s.logger.Fatal().Msg("Unable to instantiate a raft FSM")
	}

	// We use exponential backoff - default configs save for MaxElapsedTime to
	// wait for leader to get elected. We want this guardrail since followers can get
	// triggered
	exponentialBackoffEngine := backoff.NewExponentialBackOff()
	exponentialBackoffEngine.MaxElapsedTime = s.ServiceConfig.RaftTimeout

	if s.ServiceConfig.RaftLeader {
		raftConfig := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      config.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		bootStrapFut := s.raft.BootstrapCluster(raftConfig)
		if bootStrapFut.Error() != nil {
			s.logger.Fatal().Msg("Unable to bootstrap raft cluster with leader")
		}

		// readiness checker for the leader.
		leaderReadinessChecker := func() error {

			lAddr, lID := s.raft.LeaderWithID()
			if lAddr == "" || lID == "" {
				s.logger.Info().Msg("Raft Leader election in progress...")
				return errors.New("Leader not ready!")
			}
			s.logger.Info().Msg("Leader ready!!")
			return nil
		}

		err := backoff.Retry(leaderReadinessChecker, exponentialBackoffEngine)
		if err != nil {
			s.logger.Fatal().Msg("Leader not promoted yet!")
		}

	} else {
		// TODO: calling registering follower next, possibly with exponential backoff
		s.logger.Info().Msg("registering as follower ....")
		followerBody := RegisterFollowerRequest{
			FollowerId:   s.ServiceConfig.RaftNodeID,
			FollowerAddr: s.ServiceConfig.RaftAddr,
		}
		b, err := json.Marshal(followerBody)
		if err != nil {
			s.logger.Error().Msgf("Unable to unmarshal follower request. Error: %s", err)
		}

		leaderURL := fmt.Sprintf("http://%s/register-follower", s.ServiceConfig.RaftJoinAddr)
		tryJoin := func() error {
			resp, err := http.Post(leaderURL, "application/json", bytes.NewReader(b))
			if err != nil {
				s.logger.Error().Msgf("Unable to call register-follower. Got error: %s", err)
				return err
			}

			if resp.StatusCode != http.StatusOK {
				return errors.New("POST /register-follower failed... trying again")
			}
			defer resp.Body.Close()
			return nil
		}
		err = backoff.Retry(tryJoin, exponentialBackoffEngine)
		if err != nil {
			s.logger.Fatal().Msg("Unable to send register-follower to leader")
		}
		s.logger.Info().Msg("done calling register-follower on the leader")
		s.logger.Info().Msg("registration complete!!!")
	}
	s.logger.Info().Msgf("Raft Node State: %+v", s.raft.State())
	s.logger.Info().Msgf("raft stats: %+v", s.raft.Stats())

}

func (s *DKVService) Get(key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.kvmap[key]; !ok {
		return "", errors.New("key not found")
	}

	return s.kvmap[key], nil
}

func (s *DKVService) Set(key, val string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.raft.State() != raft.Leader {
		s.logger.Error().Msg("set can be done only on leader node")
		return "", errors.New("set can be done only on leader node")
	}

	if _, ok := s.kvmap[key]; ok {
		return "", errors.New("key already exists")
	}
	if s.ServiceConfig.Debug {

		s.kvmap[key] = val
		return val, nil
	}

	cmdStr := fmt.Sprintf("command:SET,key:%s,val:%s", key, val)
	applyFut := s.raft.Apply([]byte(cmdStr), s.ServiceConfig.RaftTimeout)
	err := applyFut.Error()
	if err != nil {
		return "", err
	}
	return val, nil

}

func (s *DKVService) Delete(key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.raft.State() != raft.Leader {
		s.logger.Error().Msg("set can be done only on leader node")
		return "", errors.New("set can be done only on leader node")
	}
	if _, ok := s.kvmap[key]; !ok {
		return "", errors.New("key not found")
	}

	if s.ServiceConfig.Debug {
		delete(s.kvmap, key)
		return key, nil
	}

	cmdStr := fmt.Sprintf("command:DEL,key:%s", key)
	applyFut := s.raft.Apply([]byte(cmdStr), s.ServiceConfig.RaftTimeout)
	err := applyFut.Error()
	if err != nil {
		return "", err
	}
	return key, nil

}

func (s *DKVService) RegisterFollower(followerId, followerAddr string) error {
	// get raft configs
	confFuture := s.raft.GetConfiguration()
	err := confFuture.Error()
	if err != nil {
		s.logger.Error().Msgf("Unable to get raft conf as follower: %q", err)
		return err
	}

	raftServers := confFuture.Configuration().Servers
	if len(raftServers) == 0 {
		s.logger.Fatal().Msg("No servers found. please initalize the leader first!")
	}

	_, leaderId := s.raft.LeaderWithID()
	if leaderId == "" {
		s.logger.Error().Msg("no leader in raft cluster yet")
		return errors.New("no leader in raft cluster yet")
	}

	for _, rServer := range raftServers {
		if rServer.ID == raft.ServerID(followerId) && rServer.Address == raft.ServerAddress(followerAddr) {
			// removeFuture := s.raft.RemoveServer(raft.ServerID(followerId), 0, s.ServiceConfig.RaftTimeout)
			// err := removeFuture.Error()
			// if err != nil {
			// 	s.logger.Error().
			// 		Msgf("Unable to remove existing server from raft config. Addr: %q NodeId: %q",
			// 			s.ServiceConfig.RaftAddr, s.ServiceConfig.RaftNodeID)
			// }
			// return err
			s.logger.Info().Msgf("NodeID %s already present in raft config... no need to re-register.", followerId)
			return nil
		}
	}

	// If not present in config
	// now we are clear to add this new raft server to the mix!
	addFuture := s.raft.AddVoter(
		raft.ServerID(followerId),
		raft.ServerAddress(followerAddr),
		0, s.ServiceConfig.RaftTimeout,
	)

	err = addFuture.Error()
	if err != nil {
		s.logger.Error().Msgf(
			"Unable to add server to the raft config. Node: %s Addr %s",
			s.ServiceConfig.RaftNodeID, s.ServiceConfig.RaftAddr)
		return err
	}
	s.logger.Info().Msgf("Follower registered. FollowerID: %s FollowerAddr: %s", followerId, followerAddr)
	return nil

}

func (s *DKVService) PrintConfigs() {
	s.logger.Info().Msg("--- KVService Config ---")
	s.logger.Info().Msgf("%+v", s.ServiceConfig)
	s.logger.Info().Msg("--- KVService Config ---")

}

// FSM interface funcs
func (s *DKVService) Apply(log *raft.Log) any {
	cmd := string(log.Data)
	segments := ExtractCmdSegments(cmd)
	cmdLabel := ExtractCommand(segments[0])
	key := ExtractKey(segments[1])

	// Del does not have val
	var val string
	if len(segments) == 3 {
		val = ExtractVal(segments[2])
	}

	switch cmdLabel {
	case "SET":
		s.kvmap[key] = val
	case "DEL":
		delete(s.kvmap, key)
	default:
		s.logger.Error().Msg("Unknown command label. Only SET and DEL are supported")
		return errors.New("unknown command label. Apply failed")
	}

	return nil
}

func (s *DKVService) Snapshot() (raft.FSMSnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	clonedMap := maps.Clone(s.kvmap)

	return &snapshot{kvmap: clonedMap}, nil

}
func (s *DKVService) Restore(snapshot io.ReadCloser) error {
	reconstructedKVMap := make(map[string]string)
	decoder := json.NewDecoder(snapshot)

	if err := decoder.Decode(&reconstructedKVMap); err != nil {
		s.logger.Error().Msg("Unable to restore map from snapshot")
		return err
	}
	s.kvmap = reconstructedKVMap

	return nil
}

func ExtractCmdSegments(cmd string) []string {
	segments := strings.Split(cmd, ",")
	return segments
}

func ExtractCommand(cmdSegment string) string {
	segments := strings.Split(cmdSegment, ":")
	return segments[1]
}

func ExtractKey(cmdSegment string) string {
	segments := strings.Split(cmdSegment, ":")
	return segments[1]
}

func ExtractVal(cmdSegment string) string {
	segments := strings.Split(cmdSegment, ":")
	return segments[1]
}

type snapshot struct {
	kvmap map[string]string
}

func (snap *snapshot) Persist(sink raft.SnapshotSink) error {

	// Encode data.
	b, err := json.Marshal(snap.kvmap)
	if err != nil {
		return err
	}

	// Write data to sink.
	if _, err := sink.Write(b); err != nil {
		return err
	}

	// Close the sink.
	err = sink.Close()

	if err != nil {
		sink.Cancel()
		return err
	}

	return nil
}

func (snap *snapshot) Release() {}
