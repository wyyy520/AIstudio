module github.com/aistudio/backend/tests/e2e

go 1.25.0

require (
	github.com/aistudio/backend v0.0.0
	github.com/gin-gonic/gin v1.9.1
	github.com/glebarez/sqlite v1.10.0
	github.com/google/uuid v1.6.0
	gorm.io/gorm v1.25.7
)

replace github.com/aistudio/backend => ../../apps/backend
