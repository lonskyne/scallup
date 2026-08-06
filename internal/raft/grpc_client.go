package raft

import (
	"context"

	"github.com/lonskyne/scallup/pkg/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RaftClient struct {
	client pb.RaftClient
}

func NewRaftClient(addr string) (*RaftClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &RaftClient{client: pb.NewRaftClient(conn)}, nil
}

func (rs *RaftClient) RequestVote(ctx context.Context, req *pb.RequestVoteRequest) (*pb.RequestVoteResponse, error) {
	return rs.client.RequestVote(ctx, req)
}

func (rs *RaftClient) AppendEntries(ctx context.Context, req *pb.AppendEntriesRequest) (*pb.AppendEntriesResponse, error) {
	return rs.client.AppendEntries(ctx, req)
}
