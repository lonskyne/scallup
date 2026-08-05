package raft

import (
	"context"

	"github.com/lonskyne/scallup/pkg/pb"
)

type RaftServer struct {
	pb.UnimplementedRaftServer
}

func NewRaftServer() *RaftServer {
	return &RaftServer{}
}

func (rs *RaftServer) RequestVote(ctx context.Context, req *pb.RequestVoteRequest) (*pb.RequestVoteResponse, error) {
	return &pb.RequestVoteResponse {
		Term: 0,
		VoteGranted: false,
	}, nil
}

func (rs *RaftServer) AppendEntries(ctx context.Context, req *pb.AppendEntriesRequest) (*pb.AppendEntriesResponse, error) {
	return &pb.AppendEntriesResponse {
		Term: 0,
		Success: false,
	}, nil
}
