package main

import (
	"fmt"
	"log"
	"net/http"

	"practics/internal/api"
	raftnode "practics/internal/raft" // Updated Import
)

func main() {
	// Initialize Raft Node
	raftNode, err := raftnode.NewNode() // Use raftnode instead of raft
	if err != nil {
		log.Fatalf("Failed to initialize Raft node: %v", err)
	}
	defer raftNode.Shutdown()

	// Pass raftNode to API setup
	router := api.SetupRoutes(raftNode)

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	fmt.Println("Node is running on port 8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
