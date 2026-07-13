package agent

import (
	"context"
	"fmt"

	"github.com/aistudio/packages/skill"
	"github.com/aistudio/packages/workflow"
	"github.com/google/uuid"
)

type ToolContext struct {
	Workflow *workflow.Workflow
}

type ToolFunc func(ctx context.Context, toolCtx *ToolContext, params map[string]any) (any, error)

type ToolDef struct {
	Name        string
	Description string
	Execute     ToolFunc
}

func NewToolDef(name, description string, execute ToolFunc) ToolDef {
	return ToolDef{
		Name:        name,
		Description: description,
		Execute:     execute,
	}
}

func CreateNodeTool() ToolDef {
	return NewToolDef("create_node", "Create a new node in the workflow", func(ctx context.Context, toolCtx *ToolContext, params map[string]any) (any, error) {
		nodeType, _ := params["type"].(string)
		if nodeType == "" {
			return nil, fmt.Errorf("node type is required")
		}

		nodeName, _ := params["name"].(string)
		if nodeName == "" {
			nodeName = nodeType
		}

		config, _ := params["config"].(map[string]any)
		if config == nil {
			config = make(map[string]any)
		}

		node := workflow.Node{
			ID:     uuid.New().String(),
			Type:   workflow.NodeType(nodeType),
			Name:   nodeName,
			Config: config,
			Inputs: []workflow.Port{
				{ID: "input", Name: "Input", Type: workflow.DataTypeAny},
			},
			Outputs: []workflow.Port{
				{ID: "output", Name: "Output", Type: workflow.DataTypeAny},
			},
		}

		toolCtx.Workflow.Nodes = append(toolCtx.Workflow.Nodes, node)
		return map[string]any{"node_id": node.ID, "name": nodeName}, nil
	})
}

func ConnectNodesTool() ToolDef {
	return NewToolDef("connect_nodes", "Connect two nodes in the workflow", func(ctx context.Context, toolCtx *ToolContext, params map[string]any) (any, error) {
		sourceID, _ := params["source_id"].(string)
		if sourceID == "" {
			return nil, fmt.Errorf("source_id is required")
		}
		targetID, _ := params["target_id"].(string)
		if targetID == "" {
			return nil, fmt.Errorf("target_id is required")
		}

		sourcePort, _ := params["source_port"].(string)
		if sourcePort == "" {
			sourcePort = "output"
		}
		targetPort, _ := params["target_port"].(string)
		if targetPort == "" {
			targetPort = "input"
		}

		edge := workflow.Edge{
			ID: uuid.New().String(),
			Source: workflow.EdgeEndpoint{
				NodeID: sourceID,
				PortID: sourcePort,
			},
			Target: workflow.EdgeEndpoint{
				NodeID: targetID,
				PortID: targetPort,
			},
		}

		toolCtx.Workflow.Edges = append(toolCtx.Workflow.Edges, edge)
		return map[string]any{"edge_id": edge.ID}, nil
	})
}

func FillConfigTool() ToolDef {
	return NewToolDef("fill_config", "Fill configuration parameters for a node", func(ctx context.Context, toolCtx *ToolContext, params map[string]any) (any, error) {
		nodeID, _ := params["node_id"].(string)
		if nodeID == "" {
			return nil, fmt.Errorf("node_id is required")
		}

		config, _ := params["config"].(map[string]any)
		if config == nil {
			return nil, fmt.Errorf("config is required")
		}

		for i := range toolCtx.Workflow.Nodes {
			if toolCtx.Workflow.Nodes[i].ID == nodeID {
				if toolCtx.Workflow.Nodes[i].Config == nil {
					toolCtx.Workflow.Nodes[i].Config = make(map[string]any)
				}
				for k, v := range config {
					toolCtx.Workflow.Nodes[i].Config[k] = v
				}
				return map[string]any{"node_id": nodeID, "updated": true}, nil
			}
		}

		return nil, fmt.Errorf("node not found: %s", nodeID)
	})
}

func ValidateWorkflowTool() ToolDef {
	return NewToolDef("validate_workflow", "Validate the current workflow structure", func(ctx context.Context, toolCtx *ToolContext, params map[string]any) (any, error) {
		result := workflow.ValidateWorkflow(toolCtx.Workflow)
		return map[string]any{
			"valid":    result.Valid,
			"errors":   result.Errors,
			"warnings": result.Warnings,
		}, nil
	})
}

func ApplySkillTool(skillManager *skill.SkillManager) ToolDef {
	return NewToolDef("apply_skill", "Apply a pre-built skill template to generate a workflow", func(ctx context.Context, toolCtx *ToolContext, params map[string]any) (any, error) {
		skillID, _ := params["skill_id"].(string)
		if skillID == "" {
			return nil, fmt.Errorf("skill_id is required")
		}

		s := skillManager.Get(skillID)
		if s == nil {
			return nil, fmt.Errorf("skill not found: %s", skillID)
		}

		if s.Workflow != nil {
			toolCtx.Workflow = s.Workflow
			return map[string]any{
				"skill_id":    skillID,
				"nodes_count": len(s.Workflow.Nodes),
				"edges_count": len(s.Workflow.Edges),
			}, nil
		}

		return nil, fmt.Errorf("skill %s has no workflow template", skillID)
	})
}

func SearchSkillsTool(skillManager *skill.SkillManager) ToolDef {
	return NewToolDef("search_skills", "Search for available skills", func(ctx context.Context, toolCtx *ToolContext, params map[string]any) (any, error) {
		query, _ := params["query"].(string)
		results := skillManager.Search(query)
		skills := make([]map[string]any, len(results))
		for i, s := range results {
			skills[i] = map[string]any{
				"id":          s.ID,
				"name":        s.Name,
				"description": s.Description,
				"category":    s.Category,
				"version":     s.Version,
			}
		}
		return map[string]any{"skills": skills, "count": len(skills)}, nil
	})
}

type ToolRegistry struct {
	tools map[string]ToolDef
}

func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]ToolDef),
	}
}

func (r *ToolRegistry) Register(t ToolDef) {
	r.tools[t.Name] = t
}

func (r *ToolRegistry) Get(name string) (ToolDef, bool) {
	t, ok := r.tools[name]
	return t, ok
}

func (r *ToolRegistry) List() []ToolDef {
	result := make([]ToolDef, 0, len(r.tools))
	for _, t := range r.tools {
		result = append(result, t)
	}
	return result
}