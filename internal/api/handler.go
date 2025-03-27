// internal/api/handlers.go
package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hashicorp/raft"

	"practics/internal/models"
	raftnode "practics/internal/raft" // Rename the import alias
)

type Handlers struct {
	raftServer *raft.Raft
	store      *raftnode.RaftStore
}

func NewHandlers(raftServer *raft.Raft, store *raftnode.RaftStore) *Handlers {
	return &Handlers{
		raftServer: raftServer,
		store:      store,
	}
}

func (h *Handlers) CreatePrinter(c *gin.Context) {
	var printer models.Printer
	if err := c.ShouldBindJSON(&printer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ensure unique ID
	if printer.ID == "" {
		printer.ID = generateUniqueID()
	}

	// Serialize printer
	payload, err := printer.Serialize()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize printer"})
		return
	}

	// Create Raft command
	cmd := raftnode.Command{
		Type:    raftnode.AddPrinter,
		Payload: payload,
	}
	cmdBytes, err := json.Marshal(cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create command"})
		return
	}

	// Apply to Raft
	future := h.raftServer.Apply(cmdBytes, raftApplyTimeout)
	if err := future.Error(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, printer)
}

func (h *Handlers) GetPrinters(c *gin.Context) {
	printers := h.store.GetPrinters()
	c.JSON(http.StatusOK, printers)
}

func (h *Handlers) CreateFilament(c *gin.Context) {
	var filament models.Filament
	if err := c.ShouldBindJSON(&filament); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ensure unique ID and set initial remaining weight
	if filament.ID == "" {
		filament.ID = generateUniqueID()
	}
	if filament.RemainingWeightInGrams == 0 {
		filament.RemainingWeightInGrams = filament.TotalWeightInGrams
	}

	// Serialize filament
	payload, err := filament.Serialize()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize filament"})
		return
	}

	// Create Raft command
	cmd := raftnode.Command{
		Type:    raftnode.AddFilament,
		Payload: payload,
	}
	cmdBytes, err := json.Marshal(cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create command"})
		return
	}

	// Apply to Raft
	future := h.raftServer.Apply(cmdBytes, raftApplyTimeout)
	if err := future.Error(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, filament)
}

func (h *Handlers) GetFilaments(c *gin.Context) {
	filaments := h.store.GetFilaments()
	c.JSON(http.StatusOK, filaments)
}

func (h *Handlers) CreatePrintJob(c *gin.Context) {
	var printJob models.PrintJob
	if err := c.ShouldBindJSON(&printJob); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default status to Queued
	printJob.Status = models.Queued

	// Ensure unique ID
	if printJob.ID == "" {
		printJob.ID = generateUniqueID()
	}

	// Serialize print job
	payload, err := printJob.Serialize()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize print job"})
		return
	}

	// Create Raft command
	cmd := raftnode.Command{
		Type:    raftnode.AddPrintJob,
		Payload: payload,
	}
	cmdBytes, err := json.Marshal(cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create command"})
		return
	}

	// Apply to Raft
	future := h.raftServer.Apply(cmdBytes, raftApplyTimeout)
	if err := future.Error(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, printJob)
}

func (h *Handlers) GetPrintJobs(c *gin.Context) {
	printJobs := h.store.GetPrintJobs()
	c.JSON(http.StatusOK, printJobs)
}

func (h *Handlers) UpdatePrintJobStatus(c *gin.Context) {
	jobID := c.Param("jobID")
	var statusUpdate struct {
		Status models.PrintJobStatus `json:"status"`
	}

	if err := c.ShouldBindJSON(&statusUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create Raft command for status update
	cmd := raftnode.Command{
		Type: raftnode.UpdatePrintJob,
		Payload: []byte(fmt.Sprintf(`{
			"id": "%s",
			"status": "%s"
		}`, jobID, statusUpdate.Status)),
	}
	cmdBytes, err := json.Marshal(cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create command"})
		return
	}

	// Apply to Raft
	future := h.raftServer.Apply(cmdBytes, raftApplyTimeout)
	if err := future.Error(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Print job status updated successfully"})
}

// Helper function to generate unique IDs
func generateUniqueID() string {
	return uuid.New().String()
}
