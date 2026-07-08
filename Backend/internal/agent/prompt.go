package agent

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Message represents a single message in a conversation.
type Message struct {
	Role    string `json:"role"`    // "system", "user", "assistant", "tool"
	Content string `json:"content"`
}

// LLMResponse is the structured response from an LLM.
type LLMResponse struct {
	Content string `json:"content"`
	Raw     string `json:"raw"`
}

// LLMProvider defines the interface for LLM integration.
type LLMProvider interface {
	Chat(ctx ChatContext, messages []Message) (*LLMResponse, error)
}

// ChatContext carries optional metadata for an LLM call.
type ChatContext struct {
	ProjectID string
	UserID    string
}

// ---- Prompt Templates ----

const systemPrompt = `You are AIStudio Agent, an AI assistant that helps users complete AI tasks.
You can understand natural language requests and break them down into actionable steps.

Your capabilities include:
1. Creating workflows for AI model training, data processing, and analysis
2. Installing plugins for specific AI tasks
3. Checking the environment status (Python, CUDA, dependencies)
4. Running workflows and monitoring their progress
5. Querying task status

When a user describes what they want, you should:
1. Understand their goal
2. Plan the necessary steps
3. Respond with a clear action plan in JSON format

Always respond with a JSON object containing:
- "goal": A short description of the user's goal
- "explanation": A human-readable explanation of what you'll do
- "steps": An array of actions to take, each with:
  - "tool": The tool name to call
  - "params": Parameters for the tool
  - "reason": Why this step is needed

Available tools:
{{.Tools}}

Current context:
- Project ID: {{.ProjectID}}
- Available plugins: {{.Plugins}}
- Environment status: {{.Environment}}

Respond ONLY with valid JSON, no other text.`

const plannerPrompt = `You are an AI workflow planner. Given a user's requirement, determine:
1. What type of AI task this is (vision, NLP, timeseries, etc.)
2. What plugins are needed
3. What workflow structure should be generated
4. What parameters need to be configured

Respond with a JSON object:
{
  "task_type": "vision|nlp|timeseries|simulation|system",
  "required_plugins": ["plugin_name"],
  "workflow_name": "suggested name",
  "nodes": [
    {"type": "node_type", "plugin": "plugin_name", "params": {}}
  ]
}`

const workflowGenPrompt = `Generate a complete AIStudio workflow JSON for the following requirement:
{{.Requirement}}

Available plugins and their node types:
{{.Plugins}}

The workflow JSON must follow this structure:
{
  "name": "workflow name",
  "description": "workflow description",
  "nodes": [
    {
      "id": "node_unique_id",
      "type": "node_type",
      "plugin": "plugin_name",
      "label": "Node Label",
      "params": {}
    }
  ],
  "edges": [
    {
      "id": "edge_id",
      "source": {"node_id": "source_node_id", "port_id": "output"},
      "target": {"node_id": "target_node_id", "port_id": "input"}
    }
  ]
}

Respond ONLY with the workflow JSON, no other text.`

// ---- Prompt Builder ----

// PromptBuilder constructs prompts for the LLM.
type PromptBuilder struct {
	tools       []ToolInfo
	plugins     []string
	environment string
	projectID   string
}

// NewPromptBuilder creates a new prompt builder.
func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{}
}

// SetTools sets the available tool descriptions.
func (b *PromptBuilder) SetTools(tools []ToolInfo) *PromptBuilder {
	b.tools = tools
	return b
}

// SetPlugins sets the available plugin names.
func (b *PromptBuilder) SetPlugins(plugins []string) *PromptBuilder {
	b.plugins = plugins
	return b
}

// SetEnvironment sets the environment status string.
func (b *PromptBuilder) SetEnvironment(env string) *PromptBuilder {
	b.environment = env
	return b
}

// SetProjectID sets the current project ID.
func (b *PromptBuilder) SetProjectID(projectID string) *PromptBuilder {
	b.projectID = projectID
	return b
}

// BuildSystemPrompt renders the system prompt with current context.
func (b *PromptBuilder) BuildSystemPrompt() string {
	toolsJSON, _ := json.MarshalIndent(b.tools, "", "  ")
	plugins := strings.Join(b.plugins, ", ")
	if plugins == "" {
		plugins = "none"
	}
	env := b.environment
	if env == "" {
		env = "unknown"
	}

	return strings.NewReplacer(
		"{{.Tools}}", string(toolsJSON),
		"{{.ProjectID}}", b.projectID,
		"{{.Plugins}}", plugins,
		"{{.Environment}}", env,
	).Replace(systemPrompt)
}

// BuildPlannerPrompt renders the planner prompt.
func (b *PromptBuilder) BuildPlannerPrompt(requirement string) string {
	plugins := strings.Join(b.plugins, ", ")
	if plugins == "" {
		plugins = "none"
	}

	content := strings.NewReplacer(
		"{{.Plugins}}", plugins,
	).Replace(plannerPrompt)

	return fmt.Sprintf("User requirement: %s\n\n%s", requirement, content)
}

// BuildWorkflowGenPrompt renders the workflow generation prompt.
func (b *PromptBuilder) BuildWorkflowGenPrompt(requirement string) string {
	plugins := strings.Join(b.plugins, ", ")
	if plugins == "" {
		plugins = "none"
	}

	return strings.NewReplacer(
		"{{.Requirement}}", requirement,
		"{{.Plugins}}", plugins,
	).Replace(workflowGenPrompt)
}