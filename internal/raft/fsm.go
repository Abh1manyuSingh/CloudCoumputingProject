package raft

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"practics/internal/models"

	"github.com/hashicorp/raft"
)

// Command type constants
const (
	AddPrinter  = "ADD_PRINTER"
	AddFilament = "ADD_FILAMENT"
	AddPrintJob = "ADD_PRINT_JOB"
)

// Command represents a Raft log command
type Command struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// RaftStore implements raft.FSM (Finite State Machine)
type RaftStore struct {
	mu        sync.RWMutex
	printers  map[string]*models.Printer
	filaments map[string]*models.Filament
	printJobs map[string]*models.PrintJob
}

// NewRaftStore initializes a new RaftStore.
func NewRaftStore() *RaftStore {
	return &RaftStore{
		printers:  make(map[string]*models.Printer),
		filaments: make(map[string]*models.Filament),
		printJobs: make(map[string]*models.PrintJob),
	}
}

// Apply applies a Raft log entry to the FSM.
func (s *RaftStore) Apply(log *raft.Log) interface{} {
	var cmd Command
	if err := json.Unmarshal(log.Data, &cmd); err != nil {
		fmt.Println("Failed to unmarshal command:", err)
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	switch cmd.Type {
	case AddPrinter:
		var printer models.Printer
		if err := json.Unmarshal(cmd.Payload, &printer); err != nil {
			fmt.Println("Failed to unmarshal printer:", err)
			return nil
		}
		s.printers[printer.ID] = &printer

	case AddFilament:
		var filament models.Filament
		if err := json.Unmarshal(cmd.Payload, &filament); err != nil {
			fmt.Println("Failed to unmarshal filament:", err)
			return nil
		}
		s.filaments[filament.ID] = &filament

	case AddPrintJob:
		var job models.PrintJob
		if err := json.Unmarshal(cmd.Payload, &job); err != nil {
			fmt.Println("Failed to unmarshal print job:", err)
			return nil
		}
		s.printJobs[job.ID] = &job
	}

	return nil
}

// Snapshot creates a snapshot of the current state.
func (s *RaftStore) Snapshot() (raft.FSMSnapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	return &snapshot{data: data}, nil
}

// Restore restores state from a snapshot.
func (s *RaftStore) Restore(rc io.ReadCloser) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var store RaftStore
	if err := json.NewDecoder(rc).Decode(&store); err != nil {
		return err
	}

	s.printers = store.printers
	s.filaments = store.filaments
	s.printJobs = store.printJobs
	return nil
}

// Snapshot structure
type snapshot struct {
	data []byte
}

// Persist writes snapshot data.
func (s *snapshot) Persist(sink raft.SnapshotSink) error {
	_, err := sink.Write(s.data)
	if err != nil {
		_ = sink.Cancel()
		return err
	}
	return sink.Close()
}

// Release releases the snapshot.
func (s *snapshot) Release() {}

// GetPrinters returns all printers.
func (s *RaftStore) GetPrinters() []*models.Printer {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var printers []*models.Printer
	for _, p := range s.printers {
		printers = append(printers, p)
	}
	return printers
}

// GetFilaments returns all filaments.
func (s *RaftStore) GetFilaments() []*models.Filament {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filaments []*models.Filament
	for _, f := range s.filaments {
		filaments = append(filaments, f)
	}
	return filaments
}

// GetPrintJobs returns all print jobs.
func (s *RaftStore) GetPrintJobs() []*models.PrintJob {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var jobs []*models.PrintJob
	for _, j := range s.printJobs {
		jobs = append(jobs, j)
	}
	return jobs
}
