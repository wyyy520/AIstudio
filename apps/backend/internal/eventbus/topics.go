package eventbus

// ============================================================================
// AIStudio Event Topics
//
// All inter-module communication goes through these topics.
// Topics are organized by module/domain.
// ============================================================================

// ============================================================================
// Workflow Events
// ============================================================================
const (
	TopicWorkflowCreated  Topic = "workflow.created"
	TopicWorkflowUpdated  Topic = "workflow.updated"
	TopicWorkflowDeleted  Topic = "workflow.deleted"
	TopicWorkflowValidated Topic = "workflow.validated"
	TopicWorkflowCompiled Topic = "workflow.compiled"
)

// WorkflowEventData is the payload for workflow events.
type WorkflowEventData struct {
	WorkflowID string `json:"workflowId"`
	Name       string `json:"name"`
	Target     string `json:"target,omitempty"`
	ProjectID  string `json:"projectId,omitempty"`
	Error      string `json:"error,omitempty"`
}

// ============================================================================
// Compiler Events
// ============================================================================
const (
	TopicCompileStarted   Topic = "compile.started"
	TopicCompileCompleted Topic = "compile.completed"
	TopicCompileFailed    Topic = "compile.failed"
	TopicCompileProgress  Topic = "compile.progress"
)

// CompileEventData is the payload for compiler events.
type CompileEventData struct {
	WorkflowID  string `json:"workflowId"`
	Target      string `json:"target"`
	OutputDir   string `json:"outputDir,omitempty"`
	Progress    float64 `json:"progress,omitempty"`
	Error       string `json:"error,omitempty"`
	Duration    string `json:"duration,omitempty"`
}

// ============================================================================
// Runtime Events
// ============================================================================
const (
	TopicRuntimeStarted   Topic = "runtime.started"
	TopicRuntimePreparing Topic = "runtime.preparing"
	TopicRuntimeRunning   Topic = "runtime.running"
	TopicRuntimeCompleted Topic = "runtime.completed"
	TopicRuntimeFailed    Topic = "runtime.failed"
	TopicRuntimeStopped   Topic = "runtime.stopped"
	TopicRuntimeLog       Topic = "runtime.log"
	TopicRuntimeProgress  Topic = "runtime.progress"
)

// RuntimeEventData is the payload for runtime events.
type RuntimeEventData struct {
	RunID     string  `json:"runId"`
	TaskID    string  `json:"taskId,omitempty"`
	ProjectID string  `json:"projectId,omitempty"`
	Status    string  `json:"status"`
	Progress  float64 `json:"progress,omitempty"`
	Message   string  `json:"message,omitempty"`
	Error     string  `json:"error,omitempty"`
	Duration  string  `json:"duration,omitempty"`
}

// ============================================================================
// Runtime Bundle Install Events
// ============================================================================
const (
	TopicBundleInstallStarted   Topic = "runtime:bundle_install_started"
	TopicBundleInstallProgress  Topic = "runtime:bundle_install_progress"
	TopicBundleInstallCompleted Topic = "runtime:bundle_install_completed"
	TopicBundleInstallFailed    Topic = "runtime:bundle_install_failed"
)

// BundleInstallEventData is the payload for bundle install events.
type BundleInstallEventData struct {
	BundleName string  `json:"bundleName"`
	Version    string  `json:"version"`
	Progress   float64 `json:"progress,omitempty"`
	Message    string  `json:"message,omitempty"`
	Error      string  `json:"error,omitempty"`
	DurationMs int64   `json:"durationMs,omitempty"`
}

// ============================================================================
// Task Events
// ============================================================================
const (
	TopicTaskCreated   Topic = "task.created"
	TopicTaskStarted   Topic = "task.started"
	TopicTaskCompleted Topic = "task.completed"
	TopicTaskFailed    Topic = "task.failed"
	TopicTaskCancelled Topic = "task.cancelled"
	TopicTaskProgress  Topic = "task.progress"
)

// TaskEventData is the payload for task events.
type TaskEventData struct {
	TaskID    string  `json:"taskId"`
	Type      string  `json:"type"`
	Status    string  `json:"status"`
	Progress  float64 `json:"progress,omitempty"`
	Message   string  `json:"message,omitempty"`
	Error     string  `json:"error,omitempty"`
}

// ============================================================================
// Plugin Events
// ============================================================================
const (
	TopicPluginInstalled   Topic = "plugin.installed"
	TopicPluginUninstalled Topic = "plugin.uninstalled"
	TopicPluginUpdated     Topic = "plugin.updated"
	TopicPluginEnabled     Topic = "plugin.enabled"
	TopicPluginDisabled    Topic = "plugin.disabled"
	TopicPluginError       Topic = "plugin.error"
)

// PluginEventData is the payload for plugin events.
type PluginEventData struct {
	PluginID   string `json:"pluginId"`
	Name       string `json:"name"`
	Version    string `json:"version"`
	PluginType string `json:"pluginType"`
	Error      string `json:"error,omitempty"`
}

// ============================================================================
// Project Events
// ============================================================================
const (
	TopicProjectCreated Topic = "project.created"
	TopicProjectUpdated Topic = "project.updated"
	TopicProjectDeleted Topic = "project.deleted"
	TopicProjectOpened  Topic = "project.opened"
	TopicProjectClosed  Topic = "project.closed"
)

// ProjectEventData is the payload for project events.
type ProjectEventData struct {
	ProjectID   string `json:"projectId"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	RootPath    string `json:"rootPath,omitempty"`
	Error       string `json:"error,omitempty"`
}

// ============================================================================
// Log Events
// ============================================================================
const (
	TopicLogEntry Topic = "log.entry"
	TopicLogError Topic = "log.error"
	TopicLogWarning Topic = "log.warning"
)

// LogEventData is the payload for log events.
type LogEventData struct {
	Level      string `json:"level"`
	Message    string `json:"message"`
	Source     string `json:"source"`
	TaskID     string `json:"taskId,omitempty"`
	WorkflowID string `json:"workflowId,omitempty"`
	NodeID     string `json:"nodeId,omitempty"`
	Raw        string `json:"raw,omitempty"`
}

// ============================================================================
// Diagnostic Events
// ============================================================================
const (
	TopicDiagnosticReady    Topic = "diagnostic.ready"
	TopicDiagnosticError    Topic = "diagnostic.error"
	TopicDiagnosticFixSuggested Topic = "diagnostic.fix.suggested"
)

// DiagnosticEventData is the payload for diagnostic events.
type DiagnosticEventData struct {
	TaskID     string `json:"taskId"`
	Summary    string `json:"summary"`
	Severity   string `json:"severity"`
	NodeID     string `json:"nodeId,omitempty"`
	FixCount   int    `json:"fixCount,omitempty"`
	Error      string `json:"error,omitempty"`
}

// ============================================================================
// Agent Events
// ============================================================================
const (
	TopicAgentStarted     Topic = "agent.started"
	TopicAgentCompleted   Topic = "agent.completed"
	TopicAgentWorkflowGenerated Topic = "agent.workflow.generated"
	TopicAgentError       Topic = "agent.error"
)

// AgentEventData is the payload for agent events.
type AgentEventData struct {
	SessionID  string `json:"sessionId"`
	WorkflowID string `json:"workflowId,omitempty"`
	Message    string `json:"message,omitempty"`
	Error      string `json:"error,omitempty"`
}

// ============================================================================
// Skill Events
// ============================================================================
const (
	TopicSkillApplied Topic = "skill.applied"
	TopicSkillCreated Topic = "skill.created"
	TopicSkillDeleted Topic = "skill.deleted"
)

// SkillEventData is the payload for skill events.
type SkillEventData struct {
	SkillID    string `json:"skillId"`
	Name       string `json:"name"`
	Category   string `json:"category"`
	WorkflowID string `json:"workflowId,omitempty"`
	Error      string `json:"error,omitempty"`
}

// ============================================================================
// Environment Events
// ============================================================================
const (
	TopicEnvDetecting Topic = "environment.detecting"
	TopicEnvReady     Topic = "environment.ready"
	TopicEnvError     Topic = "environment.error"
	TopicEnvInstallingBundle Topic = "environment.bundle.installing"
	TopicEnvBundleReady      Topic = "environment.bundle.ready"
)

// EnvEventData is the payload for environment events.
type EnvEventData struct {
	BundleName string `json:"bundleName,omitempty"`
	Status     string `json:"status"`
	Progress   float64 `json:"progress,omitempty"`
	Message    string `json:"message,omitempty"`
	Error      string `json:"error,omitempty"`
}

// ============================================================================
// System Events
// ============================================================================
const (
	TopicSystemStartup    Topic = "system.startup"
	TopicSystemShutdown   Topic = "system.shutdown"
	TopicSystemConfigReload Topic = "system.config.reload"
	TopicSystemError      Topic = "system.error"
)

// SystemEventData is the payload for system events.
type SystemEventData struct {
	Component string `json:"component"`
	Action    string `json:"action"`
	Message   string `json:"message,omitempty"`
	Error     string `json:"error,omitempty"`
}