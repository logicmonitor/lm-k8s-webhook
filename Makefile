include ./Makefile.Common

GOCOVMERGE=gocovmerge
TOOLS_MODULE_DIR=./internal/tools
ALL_MODULES := $(shell find . -type f -name "go.mod" -exec dirname {} \; | sort | egrep  '^./' )

VERSION_DATE ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
VERSION_PKG ?= "github.com/logicmonitor/lm-k8s-webhook/internal/version"
LM_K8S_WEBHOOK_IMG_PREFIX ?= ghcr.io/${USER}
LM_K8S_WEBHOOK_IMG_REPO ?= lm-k8s-webhook
LM_K8S_WEBHOOK_VERSION ?= "$(shell grep -v '\#' versions.txt | grep lm-k8s-webhook | awk -F= '{print $$2}')"
LM_K8S_WEBHOOK_IMG ?= ${LM_K8S_WEBHOOK_IMG_PREFIX}/${LM_K8S_WEBHOOK_IMG_REPO}:${LM_K8S_WEBHOOK_VERSION}

CMD?=

.PHONY: docker-build
docker-build:
	docker build -t ${LM_K8S_WEBHOOK_IMG} --build-arg VERSION_PKG=${VERSION_PKG} --build-arg LM_K8S_VERSION=${LM_K8S_WEBHOOK_VERSION} --build-arg VERSION_DATE=${VERSION_DATE} .

.PHONY: docker-push
docker-push:
	docker push ${LM_K8S_WEBHOOK_IMG}

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
