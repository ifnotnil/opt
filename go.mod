module github.com/ifnotnil/opt

go 1.24

// Test dependencies. They will not be pushed downstream as indirect ones.
require (
	github.com/Masterminds/semver/v3 v3.4.0
	github.com/stretchr/testify v1.11.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
