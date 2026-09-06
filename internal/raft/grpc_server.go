package raft

import (
	"context"
	"log"

	"github.com/lonskyne/scallup/pkg/pb"
)

type RaftServer struct {
	pb.UnimplementedRaftServer

	raftNode *RaftNode
}

func NewRaftServer(raftNode *RaftNode) *RaftServer {
	return &RaftServer{
		raftNode: raftNode,
	}
}

func (rs *RaftServer) RequestVote(ctx context.Context, req *pb.RequestVoteRequest) (*pb.RequestVoteResponse, error) {
	log.Printf("Received RequestVote rpc")

	term, voteGranted := rs.raftNode.ExecuteRequestVotesRPC(ctx, req.Term, req.CandidateId, req.LastLogIndex, req.LastLogTerm)

	return &pb.RequestVoteResponse{
		Term:        uint64(term),
		VoteGranted: voteGranted,
	}, nil
}

func (rs *RaftServer) AppendEntries(ctx context.Context, req *pb.AppendEntriesRequest) (*pb.AppendEntriesResponse, error) {
	log.Printf("Received AppendEntries rpc")

	term, success := rs.raftNode.ExecuteAppendEntriesRPC(ctx, req.Term, req.LeaderId, req.PrevLogIndex, req.PrevLogTerm, req.Entries, req.LeaderCommit)

	return &pb.AppendEntriesResponse{
		Term:    uint64(term),
		Success: success,
	}, nil
}
