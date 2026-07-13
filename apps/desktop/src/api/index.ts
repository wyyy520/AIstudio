/**
 * AIStudio Frontend API Module
 *
 * All API access should go through this directory.
 * Consistent structure:
 *  - client.ts:   Unified API client (all requests go through here)
 *  - health.ts:   Health check
 *  - project.ts:  Project management
 *  - workflow.ts: Workflow management
 *  - task.ts:     Task management
 *  - plugin.ts:   Plugin management
 *  - log.ts:      Log querying
 *  - websocket.ts: WebSocket connection for real-time updates
 *  - agent.ts:    AI Agent / Chat API
 *  - auth.ts:     Authentication API
 *  - settings.ts: Settings API
 *  - tauri.ts:    Tauri desktop integration
 */

export * from './client'
export * from './health'
export * from './project'
export * from './workflow'
export * from './task'
export * from './plugin'
export * from './log'
export * from './request'
export * from './websocket'
export * from './agent'
export * from './auth'
export * from './settings'
// export * from './tauri'