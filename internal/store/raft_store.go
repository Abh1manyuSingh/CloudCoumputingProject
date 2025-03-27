package raftnode

import (
	"sync"

	"practics/internal/models"
)

// RaftStore holds the application state.
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
