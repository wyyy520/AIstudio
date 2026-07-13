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
	StreamChat(ctx ChatContext, messages []Message, callback func(chunk string)) error
	GenerateJSON(ctx ChatContext, messages []Message) (map[string]interface{}, error)
}

// ChatContext carries optional metadata for an LLM call.
type ChatContext struct {
	ProjectID string
	UserID    string
}

// ---- Prompt Templates ----

const systemPrompt = `You are AIStudio Workflow Builder, an AI assistant that helps users design and configure AI workflows.
You translate natural language requirements into structured workflow.json definitions.

Your capabilities include:
1. Creating workflows with nodes and connections
2. Adding, removing, and connecting workflow nodes
3. Filling in node configuration parameters
4. Validating workflow structure and connectivity
5. Listing available plugins and skills for workflow building
6. Checking the environment for compatibility

When a user describes what they want, you should:
1. Understand their goal
2. Plan the workflow structure (nodes, edges, plugins)
3. Provide a clear action plan in JSON format

Always respond with a JSON object containing:
- "goal": A short description of the user's goal
- "explanation": A human-readable explanation of the workflow you'll create
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
1. What type of workflow this is (vision, NLP, timeseries, simulation, etc.)
2. What plugins are needed for each stage
3. What workflow structure should be generated
4. What parameters need to be configured per node

Respond with a JSON object:
{
  "task_type": "vision|nlp|timeseries|simulation|system",
  "required_plugins": ["plugin_name"],
  "workflow_name": "suggested name",
  "workflow_goal": "what the workflow should accomplish",
  "nodes": [
    {"type": "input|process|output", "plugin": "plugin_name", "params": {}}
  ],
  "edges": [
    {"source": "node_id", "target": "node_id"}
  ]
}`

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