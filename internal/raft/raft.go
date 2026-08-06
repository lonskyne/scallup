package raft

import (
	"context"
	"log"

	"github.com/lonskyne/scallup/pkg/pb"
)

type Node struct {
	NodeID int
	Peers  []Peer
}

type Peer struct {
	ID      int
	Address string
	Client  *RaftClient
}

func NewNode(nodeID int, peers map[int]string) (*Node, error) {
	peersArr := make([]Peer, len(peers))

	i := 0
	for id, addr := range peers {
		client, err := NewRaftClient(addr)
		if err != nil {
			return nil, err
		}

		peersArr[i] = Peer{
			ID:      id,
			Address: addr,
			Client:  client,
		}

		i++
	}

	return &Node{
		NodeID: nodeID,
		Peers:  peersArr, 
	}, nil
}

func (n *Node) SendHeartbeats(ctx context.Context) {
	log.Printf("SENDING HEARTBEATS")
	for _, peer := range n.Peers {
		_, err := peer.Client.AppendEntries(ctx, &pb.AppendEntriesRequest{
			Term: 0,
			LeaderId: 0,
			PrevLogIndex: 0,
			PrevLogTerm: 0,
			Entries: nil,
		})

		if err != nil {
			log.Printf("Sending heartbeat to %d failed: %v", peer.ID, err)
		}
	}
}
