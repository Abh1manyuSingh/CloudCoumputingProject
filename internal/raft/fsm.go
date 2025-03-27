// internal/raft/fsm.go
package raftnode

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"practics/internal/models"

	"github.com/hashicorp/raft"
)

type CommandType string

const (
	AddPrinter     CommandType = "add_printer"
	AddFilament    CommandType = "add_filament"
	AddPrintJob    CommandType = "add_print_job"
	UpdatePrintJob CommandType = "update_print_job"
)

type Command struct {
	Type    CommandType `json:"type"`
	Payload []byte      `json:"payload"`
}

type RaftFSM struct {
	mu        sync.RWMutex
	printers  map[string]*models.Printer
	filaments map[string]*models.Filament
	printJobs map[string]*models.PrintJob
}

func NewRaftFSM() *RaftFSM {
	return &RaftFSM{
		printers:  make(map[string]*models.Printer),
		filaments: make(map[string]*models.Filament),
		printJobs: make(map[string]*models.PrintJob),
	}
}

func (f *RaftFSM) Apply(log *raft.Log) interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()

	var cmd Command
	if err := json.Unmarshal(log.Data, &cmd); err != nil {
		return fmt.Errorf("failed to unmarshal command: %v", err)
	}

	switch cmd.Type {
	case AddPrinter:
		var printer models.Printer
		if err := json.Unmarshal(cmd.Payload, &printer); err != nil {
			return err
		}
		f.printers[printer.ID] = &printer
	case AddFilament:
		var filament models.Filament
		if err := json.Unmarshal(cmd.Payload, &filament); err != nil {
			return err
		}
		f.filaments[filament.ID] = &filament
	case AddPrintJob:
		var printJob models.PrintJob
		if err := json.Unmarshal(cmd.Payload, &printJob); err != nil {
			return err
		}
		f.printJobs[printJob.ID] = &printJob
	case UpdatePrintJob:
		var printJob models.PrintJob
		if err := json.Unmarshal(cmd.Payload, &printJob); err != nil {
			return err
		}
		f.printJobs[printJob.ID] = &printJob
	default:
		return fmt.Errorf("unknown command type: %s", cmd.Type)
	}

	return nil
}

func (f *RaftFSM) Snapshot() (raft.FSMSnapshot, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return &FSMSnapshot{
		printers:  f.printers,
		filaments: f.filaments,
		printJobs: f.printJobs,
	}, nil
}

func (f *RaftFSM) Restore(rc io.ReadCloser) error {
	defer rc.Close()

	var snapshot struct {
		Printers  map[string]*models.Printer  `json:"printers"`
		Filaments map[string]*models.Filament `json:"filaments"`
		PrintJobs map[string]*models.PrintJob `json:"print_jobs"`
	}

	if err := json.NewDecoder(rc).Decode(&snapshot); err != nil {
		return err
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	f.printers = snapshot.Printers
	f.filaments = snapshot.Filaments
	f.printJobs = snapshot.PrintJobs

	return nil
}

type FSMSnapshot struct {
	printers  map[string]*models.Printer
	filaments map[string]*models.Filament
	printJobs map[string]*models.PrintJob
}

func (s *FSMSnapshot) Persist(sink raft.SnapshotSink) error {
	err := func() error {
		// Create a snapshot payload
		snapshot := struct {
			Printers  map[string]*models.Printer  `json:"printers"`
			Filaments map[string]*models.Filament `json:"filaments"`
			PrintJobs map[string]*models.PrintJob `json:"print_jobs"`
		}{
			Printers:  s.printers,
			Filaments: s.filaments,
			PrintJobs: s.printJobs,
		}

		// Convert snapshot to JSON
		buf, err := json.Marshal(snapshot)
		if err != nil {
			return err
		}

		// Write to sink
		if _, err := sink.Write(buf); err != nil {
			return err
		}

		return sink.Close()
	}()

	if err != nil {
		sink.Cancel()
		return err
	}

	return nil
}

func (s *FSMSnapshot) Release() {}
