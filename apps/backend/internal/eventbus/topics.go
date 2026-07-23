package eventbus

import "github.com/aistudio/packages/event"

const (
	TopicWorkflowCreated    = event.TopicWorkflowCreated
	TopicWorkflowUpdated    = event.TopicWorkflowUpdated
	TopicWorkflowDeleted    = event.TopicWorkflowDeleted
	TopicWorkflowValidated  = event.TopicWorkflowValidated
	TopicWorkflowCompiled   = event.TopicWorkflowCompiled
	TopicCompileStarted     = event.TopicCompileStarted
	TopicCompileCompleted   = event.TopicCompileCompleted
	TopicCompileFailed      = event.TopicCompileFailed
	TopicCompileProgress    = event.TopicCompileProgress
	TopicRuntimeStarted     = event.TopicRuntimeStarted
	TopicRuntimePreparing   = event.TopicRuntimePreparing
	TopicRuntimeRunning     = event.TopicRuntimeRunning
	TopicRuntimeCompleted   = event.TopicRuntimeCompleted
	TopicRuntimeFailed      = event.TopicRuntimeFailed
	TopicRuntimeStopped     = event.TopicRuntimeStopped
	TopicRuntimeLog         = event.TopicRuntimeLog
	TopicRuntimeProgress    = event.TopicRuntimeProgress
	TopicBundleInstallStarted   = event.TopicBundleInstallStarted
	TopicBundleInstallProgress  = event.TopicBundleInstallProgress
	TopicBundleInstallCompleted = event.TopicBundleInstallCompleted
	TopicBundleInstallFailed    = event.TopicBundleInstallFailed
	TopicTaskCreated        = event.TopicTaskCreated
	TopicTaskStarted        = event.TopicTaskStarted
	TopicTaskCompleted      = event.TopicTaskCompleted
	TopicTaskFailed         = event.TopicTaskFailed
	TopicTaskCancelled      = event.TopicTaskCancelled
	TopicTaskProgress       = event.TopicTaskProgress
	TopicPluginInstalled    = event.TopicPluginInstalled
	TopicPluginUninstalled  = event.TopicPluginUninstalled
	TopicPluginUpdated      = event.TopicPluginUpdated
	TopicPluginEnabled      = event.TopicPluginEnabled
	TopicPluginDisabled     = event.TopicPluginDisabled
	TopicPluginError        = event.TopicPluginError
	TopicProjectCreated     = event.TopicProjectCreated
	TopicProjectUpdated     = event.TopicProjectUpdated
	TopicProjectDeleted     = event.TopicProjectDeleted
	TopicProjectOpened      = event.TopicProjectOpened
	TopicProjectClosed      = event.TopicProjectClosed
	TopicLogEntry           = event.TopicLogEntry
	TopicLogError           = event.TopicLogError
	TopicLogWarning         = event.TopicLogWarning
	TopicDiagnosticReady    = event.TopicDiagnosticReady
	TopicDiagnosticError    = event.TopicDiagnosticError
	TopicDiagnosticFixSuggested = event.TopicDiagnosticFixSuggested
	TopicAgentStarted       = event.TopicAgentStarted
	TopicAgentCompleted     = event.TopicAgentCompleted
	TopicAgentWorkflowGenerated = event.TopicAgentWorkflowGenerated
	TopicAgentError         = event.TopicAgentError
	TopicSkillApplied       = event.TopicSkillApplied
	TopicSkillCreated       = event.TopicSkillCreated
	TopicSkillDeleted       = event.TopicSkillDeleted
	TopicEnvDetecting       = event.TopicEnvDetecting
	TopicEnvReady           = event.TopicEnvReady
	TopicEnvError           = event.TopicEnvError
	TopicEnvInstallingBundle = event.TopicEnvInstallingBundle
	TopicEnvBundleReady     = event.TopicEnvBundleReady
	TopicSystemStartup      = event.TopicSystemStartup
	TopicSystemShutdown     = event.TopicSystemShutdown
	TopicSystemConfigReload = event.TopicSystemConfigReload
	TopicSystemError        = event.TopicSystemError
)

type (
	WorkflowEventData       = event.WorkflowEventData
	CompileEventData        = event.CompileEventData
	RuntimeEventData        = event.RuntimeEventData
	BundleInstallEventData  = event.BundleInstallEventData
	TaskEventData           = event.TaskEventData
	PluginEventData         = event.PluginEventData
	ProjectEventData        = event.ProjectEventData
	LogEventData            = event.LogEventData
	DiagnosticEventData     = event.DiagnosticEventData
	AgentEventData          = event.AgentEventData
	SkillEventData          = event.SkillEventData
	EnvEventData            = event.EnvEventData
	SystemEventData         = event.SystemEventData
)