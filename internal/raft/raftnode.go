package raftnode

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"practics/internal/models"

	"github.com/hashicorp/raft"
)

// Define command types as constants
const (
	AddPrinter  = "ADD_PRINTER"
	AddFilament = "ADD_FILAMENT"
	AddPrintJob = "ADD_PRINT_JOB"
)

// Command struct defines the structure of Raft log commands
type Command struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// RaftStore holds the application state and acts as Raft's FSM.
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

// Apply applies a Raft log entry to the finite state machine.
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

// Snapshot structure for serializing state without sync.RWMutex
type storeSnapshot struct {
	Printers  map[string]*models.Printer  `json:"printers"`
	Filaments map[string]*models.Filament `json:"filaments"`
	PrintJobs map[string]*models.PrintJob `json:"print_jobs"`
}

// Snapshot returns a snapshot of the current state for Raft.
func (s *RaftStore) Snapshot() (raft.FSMSnapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	snapshotData := storeSnapshot{
		Printers:  s.printers,
		Filaments: s.filaments,
		PrintJobs: s.printJobs,
	}

	data, err := json.Marshal(snapshotData)
	if err != nil {
		return nil, err
	}

	return &snapshot{data: data}, nil
}

// Restore restores the state from a snapshot.
func (s *RaftStore) Restore(rc io.ReadCloser) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var snapshotData storeSnapshot
	if err := json.NewDecoder(rc).Decode(&snapshotData); err != nil {
		return err
	}

	s.printers = snapshotData.Printers
	s.filaments = snapshotData.Filaments
	s.printJobs = snapshotData.PrintJobs

	return nil
}

// Snapshot structure
type snapshot struct {
	data []byte
}

// Persist writes the snapshot data.
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

// GetPrinters returns a list of printers.
func (s *RaftStore) GetPrinters() []*models.Printer {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var printers []*models.Printer
	for _, p := range s.printers {
		printers = append(printers, p)
	}
	return printers
}

// GetFilaments returns a list of filaments.
func (s *RaftStore) GetFilaments() []*models.Filament {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filaments []*models.Filament
	for _, f := range s.filaments {
		filaments = append(filaments, f)
	}
	return filaments
}

// GetPrintJobs returns a list of print jobs.
func (s *RaftStore) GetPrintJobs() []*models.PrintJob {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var jobs []*models.PrintJob
	for _, j := range s.printJobs {
		jobs = append(jobs, j)
	}
	return jobs
}
