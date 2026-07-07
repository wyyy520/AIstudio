package main

import (
	"log"
	"net/http"

	"github.com/aistudio/backend/internal/api"
	"github.com/aistudio/backend/internal/workflow"
)

func main() {
	workflow.RegisterDefaultNodes()
	engine := workflow.NewDefaultEngine()
	handler := api.NewHandler(engine)

	mux := http.NewServeMux()
	handler.SetupRoutes(mux)

	addr := ":8081"
	log.Printf("AIStudio Workflow Engine starting on %s", addr)
	log.Printf("API endpoints:")
	log.Printf("  POST /api/workflow/run  - Execute a workflow")
	log.Printf("  GET  /api/workflow/nodes - List registered node types")
	log.Printf("  GET  /api/health        - Health check")

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
