package service

import (
	"errors"
	"io"
	"net"
	"os"
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
	service.initRaft()
	return service
}

func (s *DKVService) initRaft() {
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

	snapshots, err := raft.NewFileSnapshotStore(s.ServiceConfig.RaftStoreDir, 1, s.logger)
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
		leaderReadinessChecker := func() error {

			lAddr, lID := s.raft.LeaderWithID()
			if lAddr == "" || lID == "" {
				s.logger.Info().Msg("Raft Leader election in progress...")
				return errors.New("Leader not ready!")
			}
			s.logger.Info().Msg("Leader ready!!")
			return nil
		}
		exponentialBackoffEngine := backoff.NewExponentialBackOff()
		exponentialBackoffEngine.MaxElapsedTime = 15 * time.Second

		err := backoff.Retry(leaderReadinessChecker, exponentialBackoffEngine)
		if err != nil {
			s.logger.Fatal().Msg("Leader not promoted yet!")
		}

	} else {
		// validate if we have a joinaddr
		if s.ServiceConfig.RaftJoinAddr == "" {
			s.logger.Fatal().Msg("Followers need to provide joining addr of leader")
		}

		// init as follower
		confFuture := s.raft.GetConfiguration()
		err := confFuture.Error()
		if err != nil {
			s.logger.Fatal().Msgf("Unable to get raft conf as follower: %q", err)
		}

		raftServers := confFuture.Configuration().Servers
		if len(raftServers) == 0 {
			s.logger.Fatal().Msg("No servers found. please initalize the leader first!")
		}

		_, leaderId := s.raft.LeaderWithID()
		if leaderId == "" {
			s.logger.Fatal().Msg("No leader in raft cluster yet!")
		}
		for _, rServer := range raftServers {
			if rServer.ID == raft.ServerID(s.ServiceConfig.RaftAddr) {
				removeFuture := s.raft.RemoveServer(rServer.ID, 0, s.ServiceConfig.RaftTimeout)
				err := removeFuture.Error()
				if err != nil {
					s.logger.Fatal().
						Msgf("Unable to remove existing server from raft config. Addr: %q NodeId: %q",
							s.ServiceConfig.RaftAddr, s.ServiceConfig.RaftNodeID)
				}
			}
		}

		// now we are clear to add this new raft server to the mix!
		addFuture := s.raft.AddVoter(
			raft.ServerID(s.ServiceConfig.RaftNodeID),
			raft.ServerAddress(s.ServiceConfig.RaftAddr),
			0, s.ServiceConfig.RaftTimeout,
		)

		err = addFuture.Error()
		if err != nil {
			s.logger.Fatal().Msgf(
				"Unable to add server to the raft config. Node: %q Addr %q",
				s.ServiceConfig.RaftNodeID, s.ServiceConfig.RaftAddr)
		}

	}

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

	if _, ok := s.kvmap[key]; ok {
		return "", errors.New("key already exists")
	}

	s.kvmap[key] = val
	return val, nil

}

func (s *DKVService) Delete(key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.kvmap[key]; !ok {
		return "", errors.New("key not found")
	}

	delete(s.kvmap, key)
	return key, nil
}

func (s *DKVService) PrintConfigs() {
	s.logger.Info().Msg("--- KVService Config ---")
	s.logger.Info().Msgf("%+v", s.ServiceConfig)
	s.logger.Info().Msg("--- KVService Config ---")

}

// FSM interface funcs
func (s *DKVService) Apply(log *raft.Log) any {
	return nil
}

func (s *DKVService) Snapshot() (raft.FSMSnapshot, error) {
	return &raft.MockSnapshot{}, nil

}
func (s *DKVService) Restore(snapshot io.ReadCloser) error {
	return nil
}
