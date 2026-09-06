package raft

import (
	"context"
	"crypto/rand"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/lonskyne/scallup/internal/storage"
	"github.com/lonskyne/scallup/pkg/pb"
)

type Role int

const (
	Follower  Role = iota
	Candidate
	Leader
)

const minElectionTimeoutMs = 10000

type RaftNode struct {
	mu            sync.Mutex

	// Persistent state
	currentTerm   int
	votedFor      *int
	log           []storage.WALEntry

	// Volatile state
	commitIndex   int
	lastApplied   int

	// Volatile state on leaders
	nextIndex     map[int]int
	matchIndex    map[int]int
	
	nodeID        int
	role          Role
	peers         []RaftPeer
	storage       *RaftStorage

	electionTimer *time.Timer
	heartbeatStop chan struct{}
}

type RaftPeer struct {
	ID      int
	Address string
	Client  *RaftClient
}

func NewRaftNode(nodeID int, peers map[int]string, storageFilePath string, wal *storage.WAL) (*RaftNode, error) {
	peersArr := make([]RaftPeer, len(peers))

	i := 0
	for id, addr := range peers {
		client, err := NewRaftClient(addr)
		if err != nil {
			return nil, err
		}

		peersArr[i] = RaftPeer{
			ID:      id,
			Address: addr,
			Client:  client,
		}

		i++
	}

	storage := NewRaftStorage(storageFilePath)
	currentTerm, votedFor, err := storage.Load()
	if err != nil {
		return nil, err
	}

	electionTimeout, err := calculateElectionTimerTimeout()
	if err != nil {
		return nil, err
	}
	electionTimer := time.NewTimer(*electionTimeout)

	fullLog, err := wal.GetFullLog()
	if err != nil {
		return nil, err
	}

	return &RaftNode{
		mu: sync.Mutex{},

		currentTerm: currentTerm,
		votedFor:    votedFor,
		log:         fullLog,

		commitIndex: 0,
		lastApplied: 0,

		nextIndex: nil,
		matchIndex: nil,

		nodeID:  nodeID,
		role:    Follower,
		peers:   peersArr,
		storage: storage,

		electionTimer: electionTimer,
		heartbeatStop: make(chan struct{}),
	}, nil
}

func (n *RaftNode) Initialize(ctx context.Context) error {
		log.Printf("Initializing raft node...")
	
    n.mu.Lock()
    term, votedFor, err := n.storage.Load()
    if err != nil {
        n.mu.Unlock()
				return err
    }

		n.currentTerm = term
		n.votedFor = votedFor

		n.role = Follower
    n.commitIndex = 0
    n.lastApplied = 0
		n.nextIndex = make(map[int]int)
		n.matchIndex = make(map[int]int)

    n.heartbeatStop = make(chan struct{})

    n.mu.Unlock()

		err = n.resetElectionTimer()
		if err != nil {
			return err
		}

    go n.electionLoop(ctx)

		return nil
}

func calculateElectionTimerTimeout() (*time.Duration, error) {
	rand, err := rand.Int(rand.Reader, big.NewInt(minElectionTimeoutMs))
	if err != nil {
		return nil, err
	}

	timeout := time.Duration(rand.Int64() + minElectionTimeoutMs) * time.Millisecond
	
	return &timeout, nil
}

func (n *RaftNode) resetElectionTimer() error {
	timeout, err := calculateElectionTimerTimeout()
	if err != nil {
		return err
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	n.electionTimer.Reset(*timeout)

	return nil
}

func (n *RaftNode) electionLoop(ctx context.Context) {
	for {
		<- n.electionTimer.C

		n.mu.Lock()

		if n.role == Leader {
			n.mu.Unlock()
			continue
		}

		n.mu.Unlock()

		n.startElection(ctx)
	}
}

func (n *RaftNode) heartbeatLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(minElectionTimeoutMs / 3) * time.Millisecond)
	defer ticker.Stop()

	beating := true
	for beating {
		select { 
		case <- ticker.C:
			n.sendHeartbeats(ctx)

		case <- n.heartbeatStop:
			beating = false;
		}
	}
}

func (n *RaftNode) startElection(ctx context.Context) {
	log.Printf("Starting leader election...")

	n.mu.Lock()

	n.role = Candidate
	n.currentTerm++
	
	n.mu.Unlock()

	n.resetElectionTimer()

	totalVotesGranted := n.requestVotes(ctx)
	totalNodes := len(n.peers) + 1

	log.Printf("Got %d votes out of %d nodes.", totalVotesGranted, totalNodes)

	if(totalVotesGranted > (totalNodes / 2)) {
		n.becomeLeader(ctx)
	}
}

func (n *RaftNode) becomeLeader(ctx context.Context) {
	log.Printf("Becoming raft leader...")

	n.role = Leader

	lastLogIndex := len(n.log)

	for _, peer := range n.peers {
		n.nextIndex[peer.ID] = lastLogIndex
		n.matchIndex[peer.ID] = 0
	}

	go n.heartbeatLoop(ctx)
}

func (n *RaftNode) requestVotes(ctx context.Context) int {
	log.Printf("Requesting votes from %d peers...", len(n.peers))

	lastLogIndex := len(n.log) - 1

	lastLogTerm := 0
	if lastLogIndex >= 0 {
		lastLogTerm = n.log[lastLogIndex].Term
	}

	// We vote for ourselves
	totalVotesGranted := 1

	for _, peer := range n.peers {
		resp, err := peer.Client.RequestVote(ctx, &pb.RequestVoteRequest{
			Term:         uint64(n.currentTerm),
			CandidateId:  uint64(n.nodeID),
			LastLogIndex: uint64(lastLogIndex),
			LastLogTerm:  uint64(lastLogTerm),
		})

		if err != nil {
			log.Printf("Requesting vote from %d failed: %v", peer.ID, err)
			return 0
		}

		if resp.VoteGranted {
			totalVotesGranted++
		}

		n.mu.Lock()

		if resp.Term > uint64(n.currentTerm) {
			n.currentTerm = int(resp.Term)
		}

		n.mu.Unlock()
	}

	return totalVotesGranted 
}

func (n *RaftNode) sendHeartbeats(ctx context.Context) {
	log.Printf("Sending heartbeats to peers...")

	prevLogIndex := len(n.log) - 1
	prevLogTerm := 0
	if prevLogIndex >= 0 {
		prevLogTerm = n.log[prevLogIndex].Term
	}

	for _, peer := range n.peers {
		_, err := peer.Client.AppendEntries(ctx, &pb.AppendEntriesRequest{
			Term:         uint64(n.currentTerm),
			LeaderId:     uint64(n.nodeID),
			PrevLogIndex: uint64(prevLogIndex),
			PrevLogTerm:  uint64(prevLogTerm),
			Entries:      nil,
		})

		if err != nil {
			log.Printf("Sending heartbeat to %d failed: %v", peer.ID, err)
		}
	}
}

func (n *RaftNode) ExecuteAppendEntriesRPC(ctx context.Context, term uint64, leaderID uint64, prevLogIndex uint64, prevLogTerm uint64, entries []*pb.LogEntry, leaderCommit uint64) (currentTerm uint, success bool) {
	return 0, false
}

func (n *RaftNode) ExecuteRequestVotesRPC(ctx context.Context, term uint64, candidateID uint64, lastLogIndex uint64, lastLogTerm uint64) (currentTerm uint, voteGranted bool) {
	return 0, false
}
