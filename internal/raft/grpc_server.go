package raft

import (
	"context"
	"log"

	"github.com/lonskyne/scallup/pkg/pb"
)

type RaftServer struct {
	pb.UnimplementedRaftServer
}

func NewRaftServer() *RaftServer {
	return &RaftServer{}
}

func (rs *RaftServer) RequestVote(ctx context.Context, req *pb.RequestVoteRequest) (*pb.RequestVoteResponse, error) {
	log.Printf("Received RequestVote rpc")

	return &pb.RequestVoteResponse{
		Term:        0,
		VoteGranted: false,
	}, nil
}

func (rs *RaftServer) AppendEntries(ctx context.Context, req *pb.AppendEntriesRequest) (*pb.AppendEntriesResponse, error) {
	log.Printf("Received AppendEntries rpc")

	return &pb.AppendEntriesResponse{
		Term:    0,
		Success: false,
	}, nil
}
