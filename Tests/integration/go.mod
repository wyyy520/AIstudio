module github.com/aistudio/tests/integration

go 1.22.0

require (
	github.com/aistudio/compiler v0.0.0
	github.com/aistudio/event v0.0.0
	github.com/aistudio/project v0.0.0
	github.com/aistudio/runtime v0.0.0
	github.com/aistudio/workflow v0.0.0
)

require (
	github.com/aistudio/common v0.0.0
	github.com/aistudio/environment v0.0.0
	github.com/google/uuid v1.6.0
)

replace (
	github.com/aistudio/agent => ../../packages/agent
	github.com/aistudio/bundles => ../../packages/bundles
	github.com/aistudio/cloud => ../../packages/cloud
	github.com/aistudio/common => ../../packages/common
	github.com/aistudio/compiler => ../../packages/compiler
	github.com/aistudio/diagnostic => ../../packages/diagnostic
	github.com/aistudio/environment => ../../packages/environment
	github.com/aistudio/event => ../../packages/event
	github.com/aistudio/generators => ../../packages/generators
	github.com/aistudio/logger => ../../packages/logger
	github.com/aistudio/plugin => ../../packages/plugin
	github.com/aistudio/project => ../../packages/project
	github.com/aistudio/runtime => ../../packages/runtime
	github.com/aistudio/sdk => ../../packages/sdk
	github.com/aistudio/security => ../../packages/security
	github.com/aistudio/skill => ../../packages/skill
	github.com/aistudio/storage => ../../packages/storage
	github.com/aistudio/workflow => ../../packages/workflow
)
