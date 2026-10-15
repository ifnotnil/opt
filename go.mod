module github.com/ifnotnil/opt

go 1.24

// Test dependencies. They will not be pushed downstream as indirect ones.
require (
	github.com/Masterminds/semver/v3 v3.5.0
	github.com/stretchr/testify v1.12.1
)

require go.yaml.in/yaml/v3 v3.0.5 // indirect
