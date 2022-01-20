include ../Makefile.Common

LM_CONFIG_RELOADER_IMG_PREFIX ?= docker-registry.logicmonitor.net/development
LM_CONFIG_RELOADER_IMG_REPO ?= lm-reloader
LM_CONFIG_RELOADER_IMG_TAG ?= dev
LM_CONFIG_RELOADER_IMG ?= ${LM_CONFIG_RELOADER_IMG_PREFIX}/${LM_CONFIG_RELOADER_IMG_REPO}:${LM_CONFIG_RELOADER_IMG_TAG}

.PHONY: docker-build
docker-build:
	docker build -t ${LM_CONFIG_RELOADER_IMG} .

.PHONY: docker-push
docker-push:
	docker push ${LM_CONFIG_RELOADER_IMG}