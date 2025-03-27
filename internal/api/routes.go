// internal/api/routes.go
package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/raft"

	raftnode "practics/internal/raft" // ✅ Correct import
)

// Timeout for Raft apply operations
var raftApplyTimeout = 5 * time.Second

func InitializeRoutes(r *gin.Engine, raftServer *raft.Raft, store *raftnode.RaftStore) {
	handlers := NewHandlers(raftServer, store)

	// API group
	v1 := r.Group("/api/v1")
	{
		// Printer endpoints
		v1.POST("/printers", handlers.CreatePrinter)
		v1.GET("/printers", handlers.GetPrinters)

		// Filament endpoints
		v1.POST("/filaments", handlers.CreateFilament)
		v1.GET("/filaments", handlers.GetFilaments)

		// Print Job endpoints
		v1.POST("/print_jobs", handlers.CreatePrintJob)
		v1.GET("/print_jobs", handlers.GetPrintJobs)
		v1.POST("/print_jobs/:jobID/status", handlers.UpdatePrintJobStatus)
	}
}
