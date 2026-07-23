package handlers

import (
	"net/http"

	"github.com/aistudio/backend/internal/project"
	"github.com/gin-gonic/gin"
)

// ProjectHandler handles project CRUD operations via the filesystem manager.
// All projects are real directories — no virtual abstraction.
type ProjectHandler struct {
	svc *project.Manager
}

// NewProjectHandler creates a new ProjectHandler.
func NewProjectHandler(mgr *project.Manager) *ProjectHandler {
	return &ProjectHandler{svc: mgr}
}

// ============================================================================
// List
// ============================================================================

// List returns all indexed projects.
// GET /api/projects
func (h *ProjectHandler) List(c *gin.Context) {
	projects := h.svc.List()
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    projects,
	})
}

// ============================================================================
// Get
// ============================================================================

// Get returns a single project by ID.
// GET /api/projects/:id
func (h *ProjectHandler) Get(c *gin.Context) {
	project, ok := h.svc.Get(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    -1,
			"message": "project not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    project,
	})
}

// ============================================================================
// Create
// ============================================================================

// Create creates a new project directory on the filesystem.
// POST /api/projects
func (h *ProjectHandler) Create(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Target      string `json:"target"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "invalid request: " + err.Error(),
		})
		return
	}

	p, err := h.svc.Create(project.CreateOptions{
		Name:        req.Name,
		Description: req.Description,
		Target:      req.Target,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "create project: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "created",
		"data":    p,
	})
}

// ============================================================================
// Update
// ============================================================================

// Update updates an existing project's metadata.
// PUT /api/projects/:id
func (h *ProjectHandler) Update(c *gin.Context) {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Target      string `json:"target"`
		Status      string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "invalid request: " + err.Error(),
		})
		return
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Target != "" {
		updates["target"] = req.Target
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}

	p, err := h.svc.Update(c.Param("id"), updates)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "updated",
		"data":    p,
	})
}

// ============================================================================
// Delete
// ============================================================================

// Delete soft-deletes a project.
// DELETE /api/projects/:id
func (h *ProjectHandler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.Param("id")); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "deleted",
	})
}

// ============================================================================
// Open (real folder)
// ============================================================================

// Open opens any real filesystem directory as an AIStudio project.
// POST /api/projects/open
func (h *ProjectHandler) Open(c *gin.Context) {
	var req struct {
		Path string `json:"path" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "path is required",
		})
		return
	}

	p, err := h.svc.Open(req.Path)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "open project: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "opened",
		"data":    p,
	})
}

// ============================================================================
// Recent
// ============================================================================

// Recent returns the most recently opened projects.
// GET /api/projects/recent
func (h *ProjectHandler) Recent(c *gin.Context) {
	projects := h.svc.Recent(10)
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    projects,
	})
}

// ============================================================================
// Workflow I/O
// ============================================================================

// ReadWorkflow reads the workflow.json for a project.
// GET /api/projects/:id/workflow
func (h *ProjectHandler) ReadWorkflow(c *gin.Context) {
	var target interface{}
	if err := h.svc.ReadWorkflow(c.Param("id"), &target); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    target,
	})
}

// SaveWorkflow saves the workflow.json for a project.
// PUT /api/projects/:id/workflow
func (h *ProjectHandler) SaveWorkflow(c *gin.Context) {
	var body interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "invalid workflow data: " + err.Error(),
		})
		return
	}

	if err := h.svc.SaveWorkflow(c.Param("id"), body); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "saved",
	})
}

// ============================================================================
// Scan
// ============================================================================

// Scan re-indexes all projects in the default projects directory.
// POST /api/projects/scan
func (h *ProjectHandler) Scan(c *gin.Context) {
	if err := h.svc.Scan(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": err.Error(),
		})
		return
	}

	projects := h.svc.List()
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "scan completed",
		"data":    projects,
	})
}
