# require variable REPO_ROOT

## https://www.gnu.org/software/make/manual/html_node/Secondary-Expansion.html
#.SECONDEXPANSION:
## https://www.gnu.org/software/make/manual/html_node/Special-Targets.html#index-not-intermediate-targets_002c-explicit
## NOTINTERMEDIATE requires make >=4.4
.NOTINTERMEDIATE:

GO_VER := $(shell go env GOVERSION | cut -d. -f1-2)

# https://www.gnu.org/software/make/manual/make.html#Automatic-Variables
# https://www.gnu.org/software/make/manual/make.html#Prerequisite-Types

TOOLS_DIR ?= $(REPO_ROOT)/.tools
TOOLS_DB ?= $(TOOLS_DIR)/.db
TOOLS_BIN ?= $(TOOLS_DIR)/bin
export TOOLS_BIN
export PATH := $(TOOLS_BIN):$(PATH)

.PHONY: tools
tools: \
	$(TOOLS_BIN)/goimports \
	$(TOOLS_BIN)/staticcheck \
	$(TOOLS_BIN)/golangci-lint \
	$(TOOLS_BIN)/gofumpt \
	$(TOOLS_BIN)/shfmt \
	$(TOOLS_BIN)/shellcheck

.PHONY: clean-tools
clean-tools:
	rm -rf $(TOOLS_DIR)

$(TOOLS_BIN):
	@mkdir -p $(TOOLS_BIN)

$(TOOLS_DB):
	@mkdir -p $(TOOLS_DB)

# In make >= 4.4. .NOTINTERMEDIATE will do the job.
.PRECIOUS: $(TOOLS_DB)/%.ver
$(TOOLS_DB)/%.ver: | $(TOOLS_DB)
	@rm -f $(TOOLS_DB)/$(word 1,$(subst ., ,$*)).*
	@touch $(TOOLS_DB)/$*.ver

define go_install
	@echo -e "Installing \e[1;36m$(1)\e[0m@\e[1;36m$(3)\e[0m using \e[1;36m$(GO_VER)\e[0m"
	GOBIN="$(TOOLS_BIN)" CGO_ENABLED=0 go install -trimpath -ldflags '-s -w -extldflags "-static"' "$(2)@$(3)"
	@echo ""
endef

# export GOMOD=$(shell pwd)/go_tools.mod
# go get -modfile='go_tools.mod' -u ...
define go_mod_ver
$(shell cd $(REPO_ROOT)/tools && go list -m $(1) | cut -d ' ' -f2)
endef

## <go ver>
.PHONY: vet
vet:
	go vet `$(GO_PACKAGES)`
	@echo ""
## </go vet>

## <staticcheck>
# https://github.com/dominikh/go-tools/releases    https://staticcheck.io/c
STATICCHECK_MOD:=honnef.co/go/tools
STATICCHECK_VER:=$(call go_mod_ver,$(STATICCHECK_MOD))
$(TOOLS_BIN)/staticcheck: $(TOOLS_DB)/staticcheck.$(STATICCHECK_VER).$(GO_VER).ver
	$(call go_install,staticcheck,$(STATICCHECK_MOD)/cmd/staticcheck,$(STATICCHECK_VER))

.PHONY: staticcheck
staticcheck: $(TOOLS_BIN)/staticcheck
	$(TOOLS_BIN)/staticcheck -f=stylish -checks=all,-ST1000 -tests ./...
	@echo ''
## </staticcheck>

## <golangci-lint>
# https://github.com/golangci/golangci-lint/releases
GOLANGCI-LINT_MOD:=github.com/golangci/golangci-lint/v2
GOLANGCI-LINT_VER:=$(call go_mod_ver,$(GOLANGCI-LINT_MOD))
$(TOOLS_BIN)/golangci-lint: $(TOOLS_DB)/golangci-lint.$(GOLANGCI-LINT_VER).$(GO_VER).ver
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(TOOLS_BIN) $(GOLANGCI-LINT_VER)

.PHONY: golangci-lint
golangci-lint: $(TOOLS_BIN)/golangci-lint
	golangci-lint run --config $(REPO_ROOT)/.golangci.yml
	@echo ''

.PHONY: golangci-lint-fmt
golangci-lint-fmt: $(TOOLS_BIN)/golangci-lint
	golangci-lint fmt --config $(REPO_ROOT)/.golangci.yml
	@echo ''
## </golangci-lint>

## <goimports>
# https://pkg.go.dev/golang.org/x/tools?tab=versions
GOIMPORTS_MOD:=golang.org/x/tools
GOIMPORTS_VER:=$(call go_mod_ver,$(GOIMPORTS_MOD))
$(TOOLS_BIN)/goimports: $(TOOLS_DB)/goimports.$(GOIMPORTS_VER).$(GO_VER).ver
	$(call go_install,goimports,$(GOIMPORTS_MOD)/cmd/goimports,$(GOIMPORTS_VER))

.PHONY: goimports
goimports: $(TOOLS_BIN)/goimports
	$(TOOLS_BIN)/goimports -w `$(GO_FILES)`

.PHONY: goimports.display
goimports.display: $(TOOLS_BIN)/goimports
	$(TOOLS_BIN)/goimports -d `$(GO_FILES)`
## </goimports>

## <gofumpt>
# https://github.com/mvdan/gofumpt/releases
GOFUMPT_MOD:=mvdan.cc/gofumpt
GOFUMPT_VER:=$(call go_mod_ver,$(GOFUMPT_MOD))
$(TOOLS_BIN)/gofumpt: $(TOOLS_DB)/gofumpt.$(GOFUMPT_VER).$(GO_VER).ver
	$(call go_install,gofumpt,$(GOFUMPT_MOD),$(GOFUMPT_VER))

.PHONY: gofumpt
gofumpt: $(TOOLS_BIN)/gofumpt
	$(TOOLS_BIN)/gofumpt -w `$(GO_FILES)`

.PHONY: gofumpt.display
gofumpt.display:
	$(TOOLS_BIN)/gofumpt -d `$(GO_FILES)`
## </gofumpt>

## <gofmt>
.PHONY: gofmt
gofmt:
	gofmt -w `$(GO_FILES)`

.PHONY: gofmt.display
gofmt.display:
	gofmt -d `$(GO_FILES)`
## </gofmt>

## <shfmt>
# https://github.com/mvdan/sh/releases
SHFMT_MOD:=mvdan.cc/sh/v3
SHFMT_VER:=$(call go_mod_ver,$(SHFMT_MOD))
$(TOOLS_BIN)/shfmt: $(TOOLS_DB)/shfmt.$(SHFMT_VER).$(GO_VER).ver
	$(call go_install,shfmt,$(SHFMT_MOD)/cmd/shfmt,$(SHFMT_VER))

.PHONY: shfmt
shfmt: $(TOOLS_BIN)/shfmt
	$(REPO_ROOT)/scripts/foreach-script $(TOOLS_BIN)/shfmt \
		--simplify \
		--language-dialect auto \
		--case-indent \
		--indent 2 \
		--write
## </shfmt>

## <shellcheck>
# https://github.com/koalaman/shellcheck/releases
SHELLCHECK_VER := v0.10.0
$(TOOLS_BIN)/shellcheck: $(TOOLS_DB)/shellcheck.$(SHELLCHECK_VER).ver | $(TOOLS_BIN)
	@$(REPO_ROOT)/scripts/install-shellcheck --version $(SHELLCHECK_VER) --destination $(TOOLS_DIR)

.PHONY: shellcheck
shellcheck: $(TOOLS_BIN)/shellcheck
	$(REPO_ROOT)/scripts/foreach-script $(TOOLS_BIN)/shellcheck \
		--norc \
		--external-sources \
		--format=tty \
		--enable=require-variable-braces,add-default-case
## </shellcheck>
