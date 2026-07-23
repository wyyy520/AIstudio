// Package compiler provides the Domain Dispatcher for multi-domain workflow compilation.
//
// Domain Dispatcher routes compilation tasks to domain-specific adapters.
// When a workflow spans multiple engineering domains (e.g., Python ML + STM32 embedded),
// the dispatcher splits the ExecutionPlan by domain and delegates each domain to
// the appropriate generator.
//
// EngStudio.md §3.15, §16.4 — Domain Dispatch Stage
package compiler

import (
	"fmt"
	"sync"

	"github.com/aistudio/packages/workflow"
)

// ============================================================================
// Domain Dispatcher
// ============================================================================

// DomainDispatcher routes compilation tasks by engineering domain.
// Each domain (Python, MATLAB, STM32, ANSYS, etc.) has its own adapter.
type DomainDispatcher struct {
	mu       sync.RWMutex
	adapters map[string]DomainAdapter
}

// DomainAdapter handles compilation for a specific engineering domain.
type DomainAdapter interface {
	// Supports returns true if this adapter can handle the given domain.
	Supports(domain workflow.Target) bool

	// Adapt transforms a section of the ExecutionPlan for this domain.
	// Returns a domain-specific plan that the domain generator can consume.
	Adapt(plan *ExecutionPlan, domain workflow.Target) (*ExecutionPlan, error)

	// Name returns the human-readable adapter name.
	Name() string
}

// NewDomainDispatcher creates a new DomainDispatcher.
func NewDomainDispatcher() *DomainDispatcher {
	return &DomainDispatcher{
		adapters: make(map[string]DomainAdapter),
	}
}

// RegisterAdapter registers a domain adapter.
func (d *DomainDispatcher) RegisterAdapter(adapter DomainAdapter) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.adapters[adapter.Name()] = adapter
}

// Dispatch splits an ExecutionPlan by domain and delegates each domain to its adapter.
// Returns a map of domain → adapted ExecutionPlan.
func (d *DomainDispatcher) Dispatch(plan *ExecutionPlan) (map[workflow.Target]*ExecutionPlan, error) {
	if plan == nil {
		return nil, fmt.Errorf("dispatch: execution plan is nil")
	}

	domains := plan.Domains
	if len(domains) == 0 {
		// Single domain — use the source target
		domains = []workflow.Target{plan.SourceTarget}
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	// Group node plans by domain
	domainNodes := make(map[workflow.Target][]string)
	for nodeID, nodePlan := range plan.NodePlans {
		domain := workflow.Target(nodePlan.Domain)
		if domain == "" {
			domain = plan.SourceTarget
		}
		domainNodes[domain] = append(domainNodes[domain], nodeID)
	}

	// Dispatch each domain
	result := make(map[workflow.Target]*ExecutionPlan, len(domainNodes))
	for domain, nodeIDs := range domainNodes {
		// Find adapter for this domain
		adapter := d.findAdapter(domain)
		if adapter == nil {
			// No adapter needed — use plan as-is for this domain
			result[domain] = d.slicePlan(plan, domain, nodeIDs)
			continue
		}

		// Slice then adapt
		sliced := d.slicePlan(plan, domain, nodeIDs)
		adapted, err := adapter.Adapt(sliced, domain)
		if err != nil {
			return nil, fmt.Errorf("domain adapter %q failed for domain %q: %w",
				adapter.Name(), domain, err)
		}
		result[domain] = adapted
	}

	return result, nil
}

// GetDomains returns the distinct domains in an ExecutionPlan.
func (d *DomainDispatcher) GetDomains(plan *ExecutionPlan) []workflow.Target {
	seen := make(map[workflow.Target]bool)
	var domains []workflow.Target

	for _, nodePlan := range plan.NodePlans {
		domain := workflow.Target(nodePlan.Domain)
		if domain == "" {
			domain = plan.SourceTarget
		}
		if !seen[domain] {
			seen[domain] = true
			domains = append(domains, domain)
		}
	}

	if len(domains) == 0 && plan.SourceTarget != "" {
		domains = append(domains, plan.SourceTarget)
	}

	return domains
}

// findAdapter finds the first adapter that supports the given domain.
func (d *DomainDispatcher) findAdapter(domain workflow.Target) DomainAdapter {
	for _, adapter := range d.adapters {
		if adapter.Supports(domain) {
			return adapter
		}
	}
	return nil
}

// slicePlan creates a sub-plan containing only the specified nodes.
func (d *DomainDispatcher) slicePlan(original *ExecutionPlan, domain workflow.Target, nodeIDs []string) *ExecutionPlan {
	nodePlans := make(map[string]NodeExecutionPlan, len(nodeIDs))
	executionOrder := make([]string, 0, len(nodeIDs))

	for _, nodeID := range nodeIDs {
		if np, ok := original.NodePlans[nodeID]; ok {
			nodePlans[nodeID] = np
			executionOrder = append(executionOrder, nodeID)
		}
	}

	return &ExecutionPlan{
		PlanVersion:    original.PlanVersion,
		WorkflowID:     original.WorkflowID,
		WorkflowName:   original.WorkflowName,
		GeneratorID:    domain,
		SourceTarget:   original.SourceTarget,
		Domains:        []workflow.Target{domain},
		ExecutionOrder: executionOrder,
		NodePlans:      nodePlans,
		OutputDir:      original.OutputDir,
		ProjectName:    original.ProjectName,
		RuntimeReq:     original.RuntimeReq,
	}
}

// ============================================================================
// Default Domain Adapter
// ============================================================================

// DefaultDomainAdapter is a no-op adapter for domains that don't need
// special handling (pass-through).
type DefaultDomainAdapter struct {
	name   string
	domain workflow.Target
}

// NewDefaultDomainAdapter creates a pass-through domain adapter.
func NewDefaultDomainAdapter(name string, domain workflow.Target) *DefaultDomainAdapter {
	return &DefaultDomainAdapter{name: name, domain: domain}
}

func (a *DefaultDomainAdapter) Supports(domain workflow.Target) bool {
	return domain == a.domain
}

func (a *DefaultDomainAdapter) Adapt(plan *ExecutionPlan, domain workflow.Target) (*ExecutionPlan, error) {
	// Pass-through — no transformation needed
	return plan, nil
}

func (a *DefaultDomainAdapter) Name() string {
	return a.name
}
