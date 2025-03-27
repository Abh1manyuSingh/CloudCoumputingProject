// internal/raft/store.go
package raftnode

import (
	"encoding/json"
	"fmt"
	"sync"

	"raft3d/internal/models"
)

type RaftStore struct {
	mu        sync.RWMutex
	printers  map[string]*models.Printer
	filaments map[string]*models.Filament
	printJobs map[string]*models.PrintJob
}

func NewRaftStore() *RaftStore {
	return &RaftStore{
		printers:  make(map[string]*models.Printer),
		filaments: make(map[string]*models.Filament),
		printJobs: make(map[string]*models.PrintJob),
	}
}

func (s *RaftStore) AddPrinter(printer *models.Printer) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.printers[printer.ID]; exists {
		return fmt.Errorf("printer with ID %s already exists", printer.ID)
	}

	s.printers[printer.ID] = printer
	return nil
}

func (s *RaftStore) GetPrinters() []*models.Printer {
	s.mu.RLock()
	defer s.mu.RUnlock()

	printers := make([]*models.Printer, 0, len(s.printers))
	for _, printer := range s.printers {
		printers = append(printers, printer)
	}
	return printers
}

func (s *RaftStore) AddFilament(filament *models.Filament) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.filaments[filament.ID]; exists {
		return fmt.Errorf("filament with ID %s already exists", filament.ID)
	}

	s.filaments[filament.ID] = filament
	return nil
}

func (s *RaftStore) GetFilaments() []*models.Filament {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filaments := make([]*models.Filament, 0, len(s.filaments))
	for _, filament := range s.filaments {
		filaments = append(filaments, filament)
	}
	return filaments
}

func (s *RaftStore) AddPrintJob(printJob *models.PrintJob) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate printer and filament exist
	if _, printerExists := s.printers[printJob.PrinterID]; !printerExists {
		return fmt.Errorf("printer with ID %s does not exist", printJob.PrinterID)
	}

	filament, filamentExists := s.filaments[printJob.FilamentID]
	if !filamentExists {
		return fmt.Errorf("filament with ID %s does not exist", printJob.FilamentID)
	}

	// Check filament weight constraints
	if printJob.PrintWeightInGrams > filament.RemainingWeightInGrams {
		return fmt.Errorf("insufficient filament weight. Requested: %d, Available: %d",
			printJob.PrintWeightInGrams, filament.RemainingWeightInGrams)
	}

	if _, exists := s.printJobs[printJob.ID]; exists {
		return fmt.Errorf("print job with ID %s already exists", printJob.ID)
	}

	s.printJobs[printJob.ID] = printJob
	return nil
}

func (s *RaftStore) GetPrintJobs() []*models.PrintJob {
	s.mu.RLock()
	defer s.mu.RUnlock()

	printJobs := make([]*models.PrintJob, 0, len(s.printJobs))
	for _, job := range s.printJobs {
		printJobs = append(printJobs, job)
	}
	return printJobs
}

func (s *RaftStore) UpdatePrintJobStatus(jobID string, newStatus models.PrintJobStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, exists := s.printJobs[jobID]
	if !exists {
		return fmt.Errorf("print job with ID %s not found", jobID)
	}

	// Validate status transitions
	switch newStatus {
	case models.Running:
		if job.Status != models.Queued {
			return fmt.Errorf("invalid status transition. Can only move from Queued to Running")
		}
	case models.Done:
		if job.Status != models.Running {
			return fmt.Errorf("invalid status transition. Can only move from Running to Done")
		}

		// Reduce filament weight when job is done
		filament, exists := s.filaments[job.FilamentID]
		if !exists {
			return fmt.Errorf("filament for job %s not found", jobID)
		}
		filament.RemainingWeightInGrams -= job.PrintWeightInGrams
	case models.Cancelled:
		if job.Status != models.Queued && job.Status != models.Running {
			return fmt.Errorf("invalid status transition. Can only cancel from Queued or Running states")
		}
	default:
		return fmt.Errorf("invalid status: %s", newStatus)
	}

	job.Status = newStatus
	return nil
}

// Serialization methods for snapshots
func (s *RaftStore) Serialize() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data := struct {
		Printers  map[string]*models.Printer  `json:"printers"`
		Filaments map[string]*models.Filament `json:"filaments"`
		PrintJobs map[string]*models.PrintJob `json:"print_jobs"`
	}{
		Printers:  s.printers,
		Filaments: s.filaments,
		PrintJobs: s.printJobs,
	}

	return json.Marshal(data)
}

func (s *RaftStore) Deserialize(data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var snapshot struct {
		Printers  map[string]*models.Printer  `json:"printers"`
		Filaments map[string]*models.Filament `json:"filaments"`
		PrintJobs map[string]*models.PrintJob `json:"print_jobs"`
	}

	if err := json.Unmarshal(data, &snapshot); err != nil {
		return err
	}

	s.printers = snapshot.Printers
	s.filaments = snapshot.Filaments
	s.printJobs = snapshot.PrintJobs

	return nil
}
