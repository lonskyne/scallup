package raft

import (
	"encoding/json"
	"fmt"
	"os"
)

type RaftStorage struct {
	filePath string
}

type raftState struct {
	Term     int  `json:"term"`
	VotedFor *int `json:"voted_for"`
}

func NewRaftStorage(filePath string) *RaftStorage {
	return &RaftStorage{
		filePath: filePath,
	}
}

func (rs *RaftStorage) SaveTerm(term int) error {
	state, err := rs.loadState()
	if err != nil {
		return err
	}

	state.Term = term

	return rs.saveState(state)
}

func (rs *RaftStorage) SaveVotedFor(votedFor *int) error {
	state, err := rs.loadState()
	if err != nil {
		return err
	}

	state.VotedFor = votedFor

	return rs.saveState(state)
}

func (rs *RaftStorage) SaveState(term int, votedFor *int) error {
	state := raftState{
		Term:     term,
		VotedFor: votedFor,
	}

	return rs.saveState(state)
}

func (rs *RaftStorage) Load() (term int, votedFor *int, err error) {
	state, err := rs.loadState()
	if err != nil {
		return 0, nil, err
	}

	return state.Term, state.VotedFor, nil
}

func (rs *RaftStorage) loadState() (raftState, error) {
	data, err := os.ReadFile(rs.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return raftState{
				Term:     0,
				VotedFor: nil,
			}, nil
		}

		return raftState{}, fmt.Errorf("failed to read Raft state: %w", err)
	}

	var state raftState

	if err := json.Unmarshal(data, &state); err != nil {
		return raftState{}, fmt.Errorf("failed to decode Raft state: %w", err)
	}

	return state, nil
}

func (rs *RaftStorage) saveState(state raftState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to encode Raft state: %w", err)
	}

	if err := os.WriteFile(rs.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write Raft state: %w", err)
	}

	return nil
}
