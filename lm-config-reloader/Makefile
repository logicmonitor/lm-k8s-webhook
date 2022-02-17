include ../Makefile.Common

VERSION_DATE ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
VERSION_PKG ?= "github.com/logicmonitor/lm-k8s-webhook/lm-config-reloader/internal/version"
LM_CONFIG_RELOADER_IMG_PREFIX ?= ghcr.io/${USER}
LM_CONFIG_RELOADER_IMG_REPO ?= lm-reloader
LM_CONFIG_RELOADER_VERSION ?= "$(shell grep -v '\#' ../versions.txt | grep lm-reloader | awk -F= '{print $$2}')"
LM_CONFIG_RELOADER_IMG ?= ${LM_CONFIG_RELOADER_IMG_PREFIX}/${LM_CONFIG_RELOADER_IMG_REPO}:${LM_CONFIG_RELOADER_VERSION}

.PHONY: docker-build
docker-build:
	docker build -t ${LM_CONFIG_RELOADER_IMG} --build-arg VERSION_PKG=${VERSION_PKG} --build-arg LM_RELOADER_VERSION=${LM_CONFIG_RELOADER_VERSION} --build-arg VERSION_DATE=${VERSION_DATE} .

.PHONY: docker-push
docker-push:
	docker push ${LM_CONFIG_RELOADER_IMG}