include ./Makefile.Common

GOCOVMERGE=gocovmerge
TOOLS_MODULE_DIR=./internal/tools
ALL_MODULES := $(shell find . -type f -name "go.mod" -exec dirname {} \; | sort | egrep  '^./' )

LM_K8s_WEBHOOK_IMG_PREFIX ?= avadhutp123
LM_K8s_WEBHOOK_IMG_REPO ?= lm-webhook
LM_K8s_WEBHOOK_IMG_TAG ?= dev
LM_K8s_WEBHOOK_IMG ?= ${LM_K8s_WEBHOOK_IMG_PREFIX}/${LM_K8s_WEBHOOK_IMG_REPO}:${LM_K8s_WEBHOOK_IMG_TAG}

CMD?=

.PHONY: docker-build
docker-build:
	docker build -t ${LM_K8s_WEBHOOK_IMG} .

.PHONY: docker-push
docker-push:
	docker push ${LM_K8s_WEBHOOK_IMG}

.PHONY: for-all
for-all:
	@echo "running $${CMD} in root"
	@$${CMD}
	@set -e; for dir in $(ALL_MODULES); do \
	  (cd "$${dir}" && \
	  	echo "running $${CMD} in $${dir}" && \
	 	$${CMD} ); \
	done

.PHONY: gomoddownload
gomoddownload:
	@$(MAKE) for-all CMD="go mod download"

.PHONY: gotest
gotest:
	@$(MAKE) for-all CMD="make test"

.PHONY: gotest-with-cover
gotest-with-cover:
	@$(MAKE) for-all CMD="make test-with-cover"
	$(GOBIN)/$(GOCOVMERGE) $$(find . -name cover.out) > coverage.txt

.PHONY: golint
golint:
	@$(MAKE) for-all CMD="make lint"

.PHONY: gotidy
gotidy:
	$(MAKE) for-all CMD="rm -fr go.sum"
	$(MAKE) for-all CMD="go mod tidy"

.PHONY: install-tools
install-tools:
	cd $(TOOLS_MODULE_DIR) && go install github.com/golangci/golangci-lint/cmd/golangci-lint
	cd $(TOOLS_MODULE_DIR) && go install github.com/wadey/gocovmerge
